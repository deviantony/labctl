package do

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/deviantony/labctl/internal/config"
	"github.com/deviantony/labctl/types"
	"github.com/digitalocean/godo"
	"go.uber.org/zap"
)

const defaultTag = "labctl"

// Client manages droplets on DigitalOcean.
type Client struct {
	ctx    context.Context
	config config.Config
	logger *zap.SugaredLogger
	client *godo.Client
}

// NewClient creates a new DigitalOcean client.
func NewClient(ctx context.Context, cfg config.Config, logger *zap.SugaredLogger) *Client {
	if cfg.TagName == "" {
		cfg.TagName = defaultTag
	}

	return &Client{
		ctx:    ctx,
		config: cfg,
		logger: logger,
		client: godo.NewFromToken(cfg.APIToken),
	}
}

// CreateDroplet creates a new droplet and returns it.
func (c *Client) CreateDroplet(name, region, size string) (types.Droplet, error) {
	resolvedRegion := mapRegion(region)
	resolvedSize := mapSize(size)

	createReq := &godo.DropletCreateRequest{
		Name:   name,
		Region: resolvedRegion,
		Size:   resolvedSize,
		Image:  godo.DropletCreateImage{Slug: c.config.BaseImage},
		Tags:   []string{c.config.TagName},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: c.config.SSHKeyFingerprint},
		},
		Monitoring: true,
	}

	newDroplet, _, err := c.client.Droplets.Create(c.ctx, createReq)
	if err != nil {
		return types.Droplet{}, fmt.Errorf("unable to create droplet: %w", err)
	}

	_, _, err = c.client.Projects.AssignResources(c.ctx, c.config.ProjectID, newDroplet.URN())
	if err != nil {
		c.logger.Warnw("Unable to assign droplet to project", "error", err)
	}

	return types.Droplet{
		ID:     newDroplet.ID,
		Name:   name,
		Region: resolvedRegion,
		Size:   resolvedSize,
	}, nil
}

// ListDroplets lists all droplets with the configured tag.
func (c *Client) ListDroplets() ([]types.Droplet, error) {
	var droplets []types.Droplet

	opt := &godo.ListOptions{PerPage: 200}
	for {
		page, resp, err := c.client.Droplets.ListByTag(c.ctx, c.config.TagName, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to list droplets: %w", err)
		}

		for _, d := range page {
			ip, _ := d.PublicIPv4()
			if ip == "" {
				ip = "-"
			}

			droplets = append(droplets, types.Droplet{
				ID:     d.ID,
				Name:   d.Name,
				Region: d.Region.Slug,
				Size:   d.Size.Slug,
				IPv4:   ip,
			})
		}

		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		currentPage, err := resp.Links.CurrentPage()
		if err != nil {
			break
		}
		opt.Page = currentPage + 1
	}

	return droplets, nil
}

// CheckAPI verifies that the API token is valid by fetching the account.
func (c *Client) CheckAPI() error {
	_, _, err := c.client.Account.Get(c.ctx)
	if err != nil {
		return fmt.Errorf("unable to reach DigitalOcean API: %w", err)
	}
	return nil
}

// RemoveDroplet deletes a droplet by ID.
func (c *Client) RemoveDroplet(id int) error {
	_, err := c.client.Droplets.Delete(c.ctx, id)
	if err != nil {
		return fmt.Errorf("unable to delete droplet %d: %w", id, err)
	}
	return nil
}

// WaitUntilReady polls until the droplet is active and SSH-reachable.
func (c *Client) WaitUntilReady(droplet *types.Droplet) error {
	deadline := time.After(c.config.PollTimeout)
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	for {
		d, _, err := c.client.Droplets.Get(c.ctx, droplet.ID)
		if err != nil {
			return err
		}

		if d.Status != "active" {
			c.logger.Infow("Waiting for droplet to be active", "status", d.Status)
		} else {
			ip, err := d.PublicIPv4()
			if err != nil {
				return err
			}

			if ip == "" {
				c.logger.Infow("Waiting for droplet to have an IP address")
			} else {
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:22", ip), 2*time.Second)
				if err != nil {
					c.logger.Infow("Waiting for SSH to be ready", "ip", ip)
				} else {
					conn.Close()
					droplet.IPv4 = ip
					return nil
				}
			}
		}

		select {
		case <-ticker.C:
			continue
		case <-deadline:
			return fmt.Errorf("timed out waiting for droplet %d to be ready", droplet.ID)
		}
	}
}

// Option represents a CLI alias mapped to a provider-specific value.
type Option struct {
	Alias string `json:"alias"`
	Slug  string `json:"slug"`
}

// RegionOptions returns all region alias-to-slug mappings.
func RegionOptions() []Option {
	return []Option{
		{"usw", "sfo3"},
		{"use", "nyc1"},
		{"eu", "fra1"},
		{"ap", "sgp1"},
		{"au", "syd1"},
	}
}

// SizeOptions returns all size alias-to-slug mappings.
func SizeOptions() []Option {
	return []Option{
		{"xs", "s-1vcpu-512mb-10gb"},
		{"s", "s-1vcpu-1gb"},
		{"m", "s-2vcpu-4gb"},
		{"l", "s-4vcpu-8gb"},
		{"xl", "s-8vcpu-16gb"},
	}
}

func mapRegion(region string) string {
	for _, o := range RegionOptions() {
		if o.Alias == region {
			return o.Slug
		}
	}
	return region
}

func mapSize(size string) string {
	for _, o := range SizeOptions() {
		if o.Alias == size {
			return o.Slug
		}
	}
	return size
}
