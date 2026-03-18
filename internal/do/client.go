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

// CreateDroplet creates a new droplet and returns it along with the action HREF for monitoring.
func (c *Client) CreateDroplet(name, region, size string) (types.Droplet, string, error) {
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

	newDroplet, resp, err := c.client.Droplets.Create(c.ctx, createReq)
	if err != nil {
		return types.Droplet{}, "", fmt.Errorf("unable to create droplet: %w", err)
	}

	// Fire-and-forget project assignment — failure is non-fatal.
	// Use a detached context so the call completes even if the parent is cancelled.
	go func() {
		ctx := context.WithoutCancel(c.ctx)
		if _, _, err := c.client.Projects.AssignResources(ctx, c.config.ProjectID, newDroplet.URN()); err != nil {
			c.logger.Warnw("Unable to assign droplet to project", "error", err)
		}
	}()

	var actionHREF string
	if resp.Links != nil && len(resp.Links.Actions) > 0 {
		actionHREF = resp.Links.Actions[0].HREF
	}

	return types.Droplet{
		ID:     newDroplet.ID,
		Name:   name,
		Region: resolvedRegion,
		Size:   resolvedSize,
	}, actionHREF, nil
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
				ID:        d.ID,
				Name:      d.Name,
				Region:    d.Region.Slug,
				Size:      d.Size.Slug,
				IPv4:      ip,
				CreatedAt: d.Created,
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
// It uses the Actions API to wait for the droplet to become active, then
// switches to a tighter polling loop for SSH readiness.
func (c *Client) WaitUntilReady(droplet *types.Droplet, actionHREF string) error {
	deadline := time.After(c.config.PollTimeout)

	// Phase 1: Wait for the droplet action to complete (active + IP).
	ip, err := c.waitForActive(droplet, actionHREF, deadline)
	if err != nil {
		return err
	}

	// Phase 2: Wait for SSH with a tighter interval (1s) since this is just a local TCP dial.
	return c.waitForSSH(droplet, ip, deadline)
}

// waitForActive polls until the droplet is active and has a public IP.
// Uses the Actions API when available, falling back to Droplets.Get.
func (c *Client) waitForActive(droplet *types.Droplet, actionHREF string, deadline <-chan time.Time) (string, error) {
	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	for {
		// Check action status if we have an action HREF.
		if actionHREF != "" {
			action, _, err := c.client.DropletActions.GetByURI(c.ctx, actionHREF)
			if err != nil {
				return "", fmt.Errorf("unable to check action status: %w", err)
			}

			if action.Status == "errored" {
				return "", fmt.Errorf("droplet creation action failed for droplet %d", droplet.ID)
			}

			if action.Status != "completed" {
				c.logger.Infow("Waiting for droplet to be active", "action_status", action.Status)
				select {
				case <-ticker.C:
					continue
				case <-deadline:
					return "", fmt.Errorf("timed out waiting for droplet %d to be ready", droplet.ID)
				}
			}

			// Action completed — no need to check it again on subsequent iterations.
			actionHREF = ""
		}

		// Action completed (or no action HREF) — fetch the droplet to get the IP.
		d, _, err := c.client.Droplets.Get(c.ctx, droplet.ID)
		if err != nil {
			return "", err
		}

		if d.Status != "active" {
			c.logger.Infow("Waiting for droplet to be active", "status", d.Status)
		} else {
			ip, err := d.PublicIPv4()
			if err != nil {
				return "", err
			}
			if ip != "" {
				return ip, nil
			}
			c.logger.Infow("Waiting for droplet to have an IP address")
		}

		select {
		case <-ticker.C:
			continue
		case <-deadline:
			return "", fmt.Errorf("timed out waiting for droplet %d to be ready", droplet.ID)
		}
	}
}

const (
	sshPollInterval = 2 * time.Second
	sshDialTimeout  = 1 * time.Second
)

// waitForSSH polls TCP port 22 on the droplet's IP with a tight interval.
func (c *Client) waitForSSH(droplet *types.Droplet, ip string, deadline <-chan time.Time) error {
	ticker := time.NewTicker(sshPollInterval)
	defer ticker.Stop()

	addr := fmt.Sprintf("%s:22", ip)

	// Check immediately before waiting for the first tick.
	if conn, err := net.DialTimeout("tcp", addr, sshDialTimeout); err == nil {
		conn.Close()
		droplet.IPv4 = ip
		return nil
	}

	c.logger.Infow("Waiting for SSH to be ready", "ip", ip)

	for {
		select {
		case <-ticker.C:
			conn, err := net.DialTimeout("tcp", addr, sshDialTimeout)
			if err != nil {
				c.logger.Infow("Waiting for SSH to be ready", "ip", ip)
				continue
			}
			conn.Close()
			droplet.IPv4 = ip
			return nil
		case <-deadline:
			return fmt.Errorf("timed out waiting for SSH on droplet %d (%s)", droplet.ID, ip)
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

var regionMap = func() map[string]string {
	m := make(map[string]string, len(RegionOptions()))
	for _, o := range RegionOptions() {
		m[o.Alias] = o.Slug
	}
	return m
}()

var sizeMap = func() map[string]string {
	m := make(map[string]string, len(SizeOptions()))
	for _, o := range SizeOptions() {
		m[o.Alias] = o.Slug
	}
	return m
}()

func mapRegion(region string) string {
	if slug, ok := regionMap[region]; ok {
		return slug
	}
	return region
}

func mapSize(size string) string {
	if slug, ok := sizeMap[size]; ok {
		return slug
	}
	return size
}
