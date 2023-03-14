package do

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/deviantony/labctl/config"
	"github.com/deviantony/labctl/types"
	"github.com/digitalocean/godo"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

type (
	VPSConfig struct {
		Name   string
		Region string
		Size   string
	}

	DOVPSBuilder struct {
		ctx    context.Context
		config config.DigitalOceanConfig
		logger *zap.SugaredLogger
		client *godo.Client
	}
)

// NewDOVPSBuilder creates a new DOVPSBuilder
func NewDOVPSBuilder(ctx context.Context, cfg config.DigitalOceanConfig, logger *zap.SugaredLogger) *DOVPSBuilder {
	client := godo.NewFromToken(cfg.APIToken)

	return &DOVPSBuilder{
		ctx:    ctx,
		config: cfg,
		logger: logger,
		client: client,
	}
}

// GetVPS retrieves information about a VPS instance based on a given ID or ID prefix
func (builder *DOVPSBuilder) GetVPS(id int) (*types.VPS, error) {
	vps, err := builder.ListVPS()
	if err != nil {
		return nil, err
	}

	matchingVPS := []types.VPS{}
	for _, v := range vps {
		if strings.HasPrefix(strconv.Itoa(v.ID), strconv.Itoa(id)) {
			matchingVPS = append(matchingVPS, v)
		}
	}

	if len(matchingVPS) == 0 {
		return nil, errors.New("no VPS found matching the given ID")
	}

	if len(matchingVPS) > 1 {
		return nil, errors.New("multiple VPS found matching the given ID, please be more specific")
	}

	return &matchingVPS[0], nil
}

// ListVPS lists all VPS instances
func (builder *DOVPSBuilder) ListVPS() ([]types.VPS, error) {
	vps := []types.VPS{}

	listOpts := &godo.ListOptions{
		PerPage: 100,
	}

	resources, _, err := builder.client.Projects.ListResources(builder.ctx, builder.config.ProjectID, listOpts)
	if err != nil {
		return vps, err
	}

	for _, resource := range resources {
		doURN := strings.Split(resource.URN, ":")
		if len(doURN) != 3 {
			builder.logger.Warnw("Skipping resource with invalid URN.",
				"URN", resource.URN,
				"Project", builder.config.ProjectID,
			)
			continue
		}

		if doURN[1] == "droplet" {
			dropletID, err := strconv.Atoi(doURN[2])
			if err != nil {
				builder.logger.Warnw("Skipping droplet with invalid identifier.",
					"URN", resource.URN,
					"dropletID", doURN[2],
					"Project", builder.config.ProjectID,
				)
				continue
			}

			droplet, _, err := builder.client.Droplets.Get(builder.ctx, dropletID)
			if err != nil {
				builder.logger.Warnw("Unable to retrieve information about a droplet.",
					"URN", resource.URN,
					"dropletID", doURN[2],
					"Project", builder.config.ProjectID,
					"error", err,
				)
				continue
			}

			for _, tag := range droplet.Tags {
				if tag == "vps" {
					v := types.VPS{
						ID:     dropletID,
						Name:   droplet.Name,
						Region: droplet.Region.Slug,
						Size:   droplet.Size.Slug,
					}

					if len(droplet.Networks.V4) > 0 {
						v.Ipv4 = droplet.Networks.V4[0].IPAddress
					} else {
						v.Ipv4 = "-"
					}

					vps = append(vps, v)
					break
				}
			}
		}
	}

	return vps, nil
}

// CreateVPS creates a new VPS instance
func (builder *DOVPSBuilder) CreateVPS(name, region, size string) (int, error) {
	config := VPSConfig{
		Name:   name,
		Region: getRegionFromOption(region),
		Size:   getSizeFromOption(size),
	}

	return builder.create(config)
}

func (builder *DOVPSBuilder) create(config VPSConfig) (int, error) {
	createRequest := &godo.DropletCreateRequest{
		Name:   config.Name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: builder.config.BaseImage,
		},
		Tags: []string{"vps"},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: builder.config.SSHKeyFingerprint},
		},
		Monitoring: true,
	}

	newDroplet, _, err := builder.client.Droplets.Create(builder.ctx, createRequest)
	if err != nil {
		builder.logger.Errorw("Unable to create droplet",
			"error", err,
		)

		return 0, err
	}

	_, _, err = builder.client.Projects.AssignResources(builder.ctx, builder.config.ProjectID, newDroplet.URN())
	if err != nil {
		builder.logger.Warnw("Unable to assign droplet to project",
			"error", err,
		)
	}

	return newDroplet.ID, nil
}

// WaitForVPS waits for a VPS instance to be ready
func (builder *DOVPSBuilder) WaitForVPSToBeReady(dropletID int) (string, error) {
	vpsIPaddr := ""

	err := wait.PollImmediate(builder.config.PollInterval, builder.config.PollTimeout,
		func() (bool, error) {
			droplet, _, err := builder.client.Droplets.Get(builder.ctx, dropletID)
			if err != nil {
				return false, err
			}

			if droplet.Status == "active" {
				publicIPV4, err := droplet.PublicIPv4()
				if err != nil {
					return false, err
				}

				if publicIPV4 == "" {
					builder.logger.Infow("Waiting for VPS to have an IP address")
					return false, nil
				}

				_, err = net.Dial("tcp", fmt.Sprintf("%s:%s", publicIPV4, "22"))
				if err != nil {
					builder.logger.Infow("Waiting for SSH service to be active",
						"IP address", publicIPV4,
					)
					return false, nil
				}

				builder.logger.Infow("VPS is ready")
				vpsIPaddr = publicIPV4
				return true, nil
			} else {
				builder.logger.Infow("Waiting for VPS to be active",
					"status", droplet.Status,
				)
				return false, nil
			}
		},
	)

	if err != nil {
		builder.logger.Errorw("Unable to poll for droplet status",
			"error", err,
		)

		return "", err
	}

	return vpsIPaddr, nil
}

// RemoveVPS removes a VPS instance
func (builder *DOVPSBuilder) RemoveVPS(dropletID int) error {
	_, err := builder.client.Droplets.Delete(builder.ctx, dropletID)
	if err != nil {
		builder.logger.Errorw("Unable to delete droplet",
			"error", err,
		)

		return err
	}

	builder.logger.Infow("Droplet successfully deleted",
		"dropletID", dropletID,
	)

	return nil
}

func getRegionFromOption(region string) string {
	switch region {
	case "usw":
		return "sfo1"
	case "use":
		return "nyc1"
	case "eu":
		return "fra1"
	case "ap":
		return "sgp1"
	case "nz":
		return "syd1"
	default:
		return ""
	}
}

func getSizeFromOption(size string) string {
	switch size {
	case "xs":
		return "s-1vcpu-512mb-10gb"
	case "s":
		return "s-1vcpu-1gb"
	case "m":
		return "s-2vcpu-4gb"
	case "l":
		return "s-4vcpu-8gb"
	case "xl":
		return "s-8vcpu-16gb"
	default:
		return ""
	}
}
