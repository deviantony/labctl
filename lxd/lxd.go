package lxd

import (
	"context"
	"errors"
	"os"

	"github.com/deviantony/labctl/config"
	"github.com/deviantony/labctl/tls"
	"github.com/deviantony/labctl/types"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"go.uber.org/zap"
)

type (
	// FlaskManager is used to manage flasks via LXD
	FlaskManager struct {
		logger *zap.SugaredLogger
		client *lxd.InstanceServer
	}
)

func fileExists(path string, logger *zap.SugaredLogger) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		logger.Errorf("An error occurred while checking if the file exists: %s", err)
		return true
	}
}

// NewFlaskManager creates a new flask manager
// It can create and manage flasks via a LXD server
func NewFlaskManager(ctx context.Context, cfg config.LXDConfig, logger *zap.SugaredLogger) (*FlaskManager, error) {
	logger.Info("Verifying TLS certificates existence")

	if !fileExists(cfg.Cert, logger) || !fileExists(cfg.Key, logger) {
		logger.Info("Unable to locate TLS certificate and key, generating new ones")

		err := tls.GenerateSelfSignedTLSCertificates(logger, cfg.Cert, cfg.Key)
		if err != nil {
			logger.Errorf("Unable to generate TLS certificates: %s", err)
			return nil, err
		}
	} else {
		logger.Info("TLS certificate and key available, skipping generation")
	}

	clientCertBytes, err := os.ReadFile(cfg.Cert)
	if err != nil {
		logger.Errorf("Unable to read client certificate: %s", err)
		return nil, err
	}

	clientKeyBytes, err := os.ReadFile(cfg.Key)
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

	logger.Infof("Connected to LXD server, auth status: %s", serverInfo.Auth)

	if serverInfo.Auth != "trusted" {
		logger.Info("Authenticating with server")

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
		client: &client,
		logger: logger,
	}, nil
}

// ListFlasks lists all the flasks running in DigitalOcean (inside a specific project)
func (manager *FlaskManager) ListFlasks() ([]types.Flask, error) {
	flasks := []types.Flask{}

	instances, err := (*manager.client).GetInstances(api.InstanceTypeContainer)
	if err != nil {
		manager.logger.Errorf("Unable to send certificate trust request: %s", err)
		return flasks, err
	}

	for _, instance := range instances {
		flasks = append(flasks, types.Flask{
			ID:     0,
			Name:   instance.Name,
			Region: "-",
			Size:   "-",
			Ipv4:   "-",
		})
	}

	return flasks, nil
}
