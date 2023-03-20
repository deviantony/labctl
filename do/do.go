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

const LABCTL_FLASK_TAG = "labctl-flask"

type (
	// FlaskManager is used to manage flasks in DigitalOcean
	FlaskManager struct {
		ctx    context.Context
		config config.DigitalOceanConfig
		logger *zap.SugaredLogger
		client *godo.Client
	}

	dropletConfig struct {
		Name   string
		Region string
		Size   string
	}
)

// NewFlaskManager creates a new flask manager
// It can create and manage flasks in DigitalOcean as droplets
func NewFlaskManager(ctx context.Context, cfg config.DigitalOceanConfig, logger *zap.SugaredLogger) *FlaskManager {
	client := godo.NewFromToken(cfg.APIToken)

	return &FlaskManager{
		ctx:    ctx,
		config: cfg,
		logger: logger,
		client: client,
	}
}

// GetFlask retrieves information about a flask based on a given ID or ID prefix
func (manager *FlaskManager) GetFlask(id int) (*types.Flask, error) {
	flasks, err := manager.ListFlasks()
	if err != nil {
		return nil, err
	}

	matches := []types.Flask{}
	for _, v := range flasks {
		if strings.HasPrefix(strconv.Itoa(v.ID), strconv.Itoa(id)) {
			matches = append(matches, v)
		}
	}

	if len(matches) == 0 {
		return nil, errors.New("no flask found matching the given ID")
	}

	if len(matches) > 1 {
		return nil, errors.New("multiple flasks found matching the given ID, please be more specific")
	}

	return &matches[0], nil
}

// ListFlasks lists all the flasks running in DigitalOcean (inside a specific project)
func (manager *FlaskManager) ListFlasks() ([]types.Flask, error) {
	flasks := []types.Flask{}

	listOpts := &godo.ListOptions{
		PerPage: 100,
	}

	resources, _, err := manager.client.Projects.ListResources(manager.ctx, manager.config.ProjectID, listOpts)
	if err != nil {
		return flasks, err
	}

	for _, resource := range resources {
		doURN := strings.Split(resource.URN, ":")
		if len(doURN) != 3 {
			manager.logger.Warnw("Skipping resource with invalid URN.",
				"URN", resource.URN,
				"Project", manager.config.ProjectID,
			)
			continue
		}

		if doURN[1] == "droplet" {
			dropletID, err := strconv.Atoi(doURN[2])
			if err != nil {
				manager.logger.Warnw("Skipping droplet with invalid identifier.",
					"URN", resource.URN,
					"dropletID", doURN[2],
					"Project", manager.config.ProjectID,
				)
				continue
			}

			droplet, _, err := manager.client.Droplets.Get(manager.ctx, dropletID)
			if err != nil {
				manager.logger.Warnw("Unable to retrieve information about a droplet.",
					"URN", resource.URN,
					"dropletID", doURN[2],
					"Project", manager.config.ProjectID,
					"error", err,
				)
				continue
			}

			for _, tag := range droplet.Tags {
				if tag == LABCTL_FLASK_TAG {
					flask := types.Flask{
						ID:     dropletID,
						Name:   droplet.Name,
						Region: droplet.Region.Slug,
						Size:   droplet.Size.Slug,
					}

					if len(droplet.Networks.V4) > 0 {
						flask.Ipv4 = droplet.Networks.V4[0].IPAddress
					} else {
						flask.Ipv4 = "-"
					}

					flasks = append(flasks, flask)
					break
				}
			}
		}
	}

	return flasks, nil
}

// CreateFlask creates a new flask as a droplet in DigitalOcean
func (manager *FlaskManager) CreateFlask(name, region, size string) (int, error) {
	config := dropletConfig{
		Name:   name,
		Region: getRegionFromOption(region),
		Size:   getSizeFromOption(size),
	}

	return manager.createDroplet(config)
}

func (manager *FlaskManager) createDroplet(config dropletConfig) (int, error) {
	createRequest := &godo.DropletCreateRequest{
		Name:   config.Name,
		Region: config.Region,
		Size:   config.Size,
		Image: godo.DropletCreateImage{
			Slug: manager.config.BaseImage,
		},
		Tags: []string{LABCTL_FLASK_TAG},
		SSHKeys: []godo.DropletCreateSSHKey{
			{Fingerprint: manager.config.SSHKeyFingerprint},
		},
		Monitoring: true,
	}

	newDroplet, _, err := manager.client.Droplets.Create(manager.ctx, createRequest)
	if err != nil {
		manager.logger.Errorw("Unable to create droplet",
			"error", err,
		)

		return 0, err
	}

	_, _, err = manager.client.Projects.AssignResources(manager.ctx, manager.config.ProjectID, newDroplet.URN())
	if err != nil {
		manager.logger.Warnw("Unable to assign droplet to project",
			"error", err,
		)
	}

	return newDroplet.ID, nil
}

// WaitUntilFlaskIsReady waits for a flask to be ready
// It returns the IP address of the flask when it is ready
func (manager *FlaskManager) WaitUntilFlaskIsReady(id int) (string, error) {
	flaskIP := ""

	err := wait.PollImmediate(manager.config.PollInterval, manager.config.PollTimeout,
		func() (bool, error) {
			droplet, _, err := manager.client.Droplets.Get(manager.ctx, id)
			if err != nil {
				return false, err
			}

			if droplet.Status == "active" {
				publicIPV4, err := droplet.PublicIPv4()
				if err != nil {
					return false, err
				}

				if publicIPV4 == "" {
					manager.logger.Infow("Waiting for flask to have an IP address")
					return false, nil
				}

				_, err = net.Dial("tcp", fmt.Sprintf("%s:%s", publicIPV4, "22"))
				if err != nil {
					manager.logger.Infow("Waiting for SSH service to be active",
						"IP address", publicIPV4,
					)
					return false, nil
				}

				manager.logger.Infow("Flask is ready")
				flaskIP = publicIPV4
				return true, nil
			} else {
				manager.logger.Infow("Waiting for flask to be active",
					"status", droplet.Status,
				)
				return false, nil
			}
		},
	)

	if err != nil {
		manager.logger.Errorw("Unable to poll for droplet status",
			"error", err,
		)

		return "", err
	}

	return flaskIP, nil
}

// RemoveFlask removes a flask
func (manager *FlaskManager) RemoveFlask(id int) error {
	_, err := manager.client.Droplets.Delete(manager.ctx, id)
	if err != nil {
		manager.logger.Errorw("Unable to delete droplet",
			"error", err,
		)

		return err
	}

	manager.logger.Infow("Flask successfully deleted",
		"dropletID", id,
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
