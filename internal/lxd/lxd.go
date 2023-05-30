package lxd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/deviantony/labctl/internal/config"
	"github.com/deviantony/labctl/pkg/filesystem"
	"github.com/deviantony/labctl/pkg/tls"
	"github.com/deviantony/labctl/types"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// TAG_FLASKID is the tag used to identify flasks
	// It is associated with all LXC containers created via this utility
	TAG_FLASKID = "user.flask-id"
	// BASE_IMAGE is the base image used to create flasks
	BASE_IMAGE = "flask-docker"
	// DEFAULT_IMAGE_SERVER is the default server used to fetch images
	DEFAULT_IMAGE_SERVER = "https://images.linuxcontainers.org"
	// LXD_PROFILE_PREFIX is the prefix used for LXD profiles
	LXD_PROFILE_PREFIX = "labctl-flask-"
	// DOCKER_STORAGE is the name of the docker storage
	DOCKER_STORAGE = "docker-data"
	// DOCKER_VOLUME_PREFIX is the prefix used for docker volumes
	DOCKER_VOLUME_PREFIX = "docker-"
	// DOCKER_VOLUME_SIZE is the size of the docker volume
	DOCKER_VOLUME_SIZE = "5GB"
)

type (
	// FlaskManager is used to manage flasks via LXD
	FlaskManager struct {
		logger *zap.SugaredLogger
		cfg    config.LXDConfig
		client lxd.InstanceServer
	}
)

// NewFlaskManager creates a new flask manager
// It can create and manage flasks via a LXD server
func NewFlaskManager(ctx context.Context, cfg config.LXDConfig, logger *zap.SugaredLogger) (*FlaskManager, error) {
	logger.Debug("Verifying TLS certificates existence")

	clientCertExists, err := filesystem.FileExists(cfg.Client.Cert)
	if err != nil {
		logger.Errorf("Unable to verify TLS certificate existence: %s", err)
		return nil, err
	}

	clientKeyExists, err := filesystem.FileExists(cfg.Client.Key)
	if err != nil {
		logger.Errorf("Unable to verify TLS key existence: %s", err)
		return nil, err
	}

	if !clientCertExists || !clientKeyExists {
		logger.Debug("Unable to locate TLS certificate and key, generating new ones")

		err := tls.GenerateSelfSignedTLSCertificates(cfg.Client.Key, cfg.Client.Cert)
		if err != nil {
			logger.Errorf("Unable to generate TLS certificates: %s", err)
			return nil, err
		}
	} else {
		logger.Debug("TLS certificate and key available, skipping generation")
	}

	clientCertBytes, err := os.ReadFile(cfg.Client.Cert)
	if err != nil {
		logger.Errorf("Unable to read client certificate: %s", err)
		return nil, err
	}

	clientKeyBytes, err := os.ReadFile(cfg.Client.Key)
	if err != nil {
		logger.Errorf("Unable to read client key: %s", err)
		return nil, err
	}

	client, err := lxd.ConnectLXDWithContext(ctx, cfg.Server.Addr, &lxd.ConnectionArgs{
		TLSClientCert:      string(clientCertBytes),
		TLSClientKey:       string(clientKeyBytes),
		InsecureSkipVerify: true,
	})
	if err != nil {
		logger.Errorf("Unable to connect to lxd: %s", err)
		return nil, err
	}

	serverInfo, _, err := client.GetServer()
	if err != nil {
		logger.Errorf("Unable to retrieve server information: %s", err)
		return nil, err
	}

	logger.Debugf("Connected to LXD server, auth status: %s", serverInfo.Auth)

	if serverInfo.Auth != "trusted" {
		logger.Debugf("Authenticating with server")

		err := client.CreateCertificate(api.CertificatesPost{
			CertificatePut: api.CertificatePut{
				Type: "client",
			},
			Password: cfg.Server.Password,
		})

		if err != nil {
			logger.Errorf("Unable to send certificate trust request: %s", err)
			return nil, err
		}
	}

	return &FlaskManager{
		client: client,
		logger: logger,
		cfg:    cfg,
	}, nil
}

// CreateFlask creates a new flask
func (manager *FlaskManager) CreateFlask(name string, cfg types.FlaskConfig) (types.Flask, error) {
	flask := types.Flask{
		Name: name,
	}

	profile := cfg.Profile
	if profile == "" {
		profile = getProfileFromSizeOption(cfg.Size)
	}

	image := cfg.Image
	if image == "" {
		image = BASE_IMAGE
	}

	err := manager.createLXDInstance(name, image, profile)
	if err != nil {
		return flask, err
	}

	dockerVolumeName := DOCKER_VOLUME_PREFIX + name
	err = manager.createLXDStorageVolume(DOCKER_STORAGE, dockerVolumeName, DOCKER_VOLUME_SIZE)
	if err != nil {
		return flask, err
	}

	err = manager.attachLXDVolumeToInstance(DOCKER_STORAGE, dockerVolumeName, name, "/var/lib/docker")
	if err != nil {
		return flask, err
	}

	sshPubKey, err := os.ReadFile(manager.cfg.SSHPublicKey)
	if err != nil {
		manager.logger.Errorw("Unable to open SSH public key file",
			"error", err,
		)

		return flask, err
	}

	err = manager.createFileInLXDInstance(name, "/root/.ssh/authorized_keys", sshPubKey)
	if err != nil {
		return flask, err
	}

	err = manager.startLXDInstance(name)
	if err != nil {
		return flask, err
	}

	return flask, nil
}

// GetFlask retrieves information about a flask based on a given ID or ID prefix
func (manager *FlaskManager) GetFlask(id string) (types.Flask, error) {
	flasks, err := manager.ListFlasks()
	if err != nil {
		return types.Flask{}, err
	}

	matches := []types.Flask{}
	for _, flask := range flasks {
		if strings.HasPrefix(flask.LXD.ID, id) {
			matches = append(matches, flask)
		}
	}

	if len(matches) == 0 {
		return types.Flask{}, errors.New("no flask found matching the given ID")
	}

	if len(matches) > 1 {
		return types.Flask{}, errors.New("multiple flasks found matching the given ID, please be more specific")
	}

	return matches[0], nil
}

// ListFlasks lists all the flasks running in DigitalOcean (inside a specific project)
func (manager *FlaskManager) ListFlasks() ([]types.Flask, error) {
	flasks := []types.Flask{}

	instances, err := manager.client.GetInstances(api.InstanceTypeContainer)
	if err != nil {
		manager.logger.Errorf("Unable to send certificate trust request: %s", err)
		return flasks, err
	}

	for _, instance := range instances {
		instanceState, err := manager.getLXDInstanceState(instance.Name)
		if err != nil {
			manager.logger.Warnw("Unable to retrieve flask state",
				"name", instance.Name,
			)
			continue
		}

		id, ok := instance.Config[TAG_FLASKID]
		if !ok {
			manager.logger.Warnw("Unable to retrieve flask ID",
				"name", instance.Name,
			)
			continue
		}

		flask := types.Flask{
			Name: instance.Name,
			LXD: types.FlaskLXDProperties{
				ID:       id,
				Profiles: instance.Profiles,
				Status:   instanceState.Status,
			},
		}

		if instanceState.Status == "Running" {
			network, ok := instanceState.Network["eth0"]
			if !ok {
				manager.logger.Warnw("Unable to retrieve network interface details for flask",
					"name", instance.Name,
				)
				continue
			}

			for _, address := range network.Addresses {
				if address.Family == "inet" {
					flask.Ipv4 = address.Address
					break
				}
			}
		} else {
			flask.Ipv4 = "-"
		}

		flasks = append(flasks, flask)
	}

	return flasks, nil
}

// RemoveFlask deletes a flask
func (manager *FlaskManager) RemoveFlask(flask types.Flask) error {
	manager.logger.Infow("Removing flask",
		"name", flask.Name,
	)

	state, err := manager.getLXDInstanceState(flask.Name)
	if err != nil {
		return err
	}

	if state.Status == "Running" {
		err := manager.stopLXDInstance(flask.Name)
		if err != nil {
			return err
		}
	}

	err = manager.removeLXDInstance(flask.Name)
	if err != nil {
		return err
	}

	time.Sleep(3 * time.Second)

	dockerVolumeName := DOCKER_VOLUME_PREFIX + flask.Name
	return manager.removeLXDStorageVolume(DOCKER_STORAGE, dockerVolumeName)
}

// WaitUntilFlaskIsReady waits until the flask is ready to be used
// We wait until the flask is running and has an IP address that is reachable
func (manager *FlaskManager) WaitUntilFlaskIsReady(flask *types.Flask) error {
	instance, err := manager.getLXDInstance(flask.Name)
	if err != nil {
		return err
	}

	flask.LXD.ID = instance.Config[TAG_FLASKID]

	return wait.PollImmediate(time.Duration(3*time.Second), manager.cfg.Client.Timeout, func() (bool, error) {
		instanceState, err := manager.getLXDInstanceState(flask.Name)
		if err != nil {
			return false, err
		}

		if instanceState.Status == "Running" {
			network, ok := instanceState.Network["eth0"]
			if !ok {
				manager.logger.Infow("Waiting for flask to have an IP address")
				return false, nil
			}

			for _, address := range network.Addresses {
				if address.Family == "inet" {
					flask.Ipv4 = address.Address

					d := net.Dialer{Timeout: 1 * time.Second}
					_, err = d.Dial("tcp", fmt.Sprintf("%s:%d", flask.Ipv4, 22))
					if err != nil {
						manager.logger.Infow("Waiting for SSH service to be ready")
						return false, nil
					}

					return true, nil
				}
			}

			manager.logger.Infow("Waiting for flask to have an IP address")
			return false, nil
		} else {
			manager.logger.Infow("Waiting for flask to be active",
				"status", instanceState.Status,
			)
			return false, nil
		}
	})
}
