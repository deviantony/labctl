package lxd

import (
	"context"
	"errors"
	"io/ioutil"
	"os"

	"github.com/deviantony/labctl/tls"
	lxd "github.com/lxc/lxd/client"
	"github.com/lxc/lxd/shared/api"
	"go.uber.org/zap"
)

const (
	LXD_SERVER_ADDR      = "https://homesrv.local:8443"
	LXD_SERVER_PASSWORD  = "rewop27"
	TLS_CERTIFICATE_PATH = "data/cert.pem"
	TLS_KEY_PATH         = "data/key.pem"
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
func NewFlaskManager(ctx context.Context, logger *zap.SugaredLogger) (*FlaskManager, error) {
	logger.Info("Verifying TLS certificates existence")

	if !fileExists(TLS_CERTIFICATE_PATH, logger) || !fileExists(TLS_KEY_PATH, logger) {
		logger.Info("Unable to locate TLS certificate and key, generating new ones")

		err := tls.GenerateSelfSignedTLSCertificates(logger, TLS_CERTIFICATE_PATH, TLS_KEY_PATH)
		if err != nil {
			logger.Errorf("Unable to generate TLS certificates: %s", err)
			return nil, err
		}
	} else {
		logger.Info("TLS certificate and key available, skipping generation")
	}

	clientCertBytes, err := ioutil.ReadFile(TLS_CERTIFICATE_PATH)
	if err != nil {
		logger.Errorf("Unable to read client certificate: %s", err)
		return nil, err
	}

	clientKeyBytes, err := ioutil.ReadFile(TLS_KEY_PATH)
	if err != nil {
		logger.Errorf("Unable to read client key: %s", err)
		return nil, err
	}

	client, err := lxd.ConnectLXDWithContext(ctx, LXD_SERVER_ADDR, &lxd.ConnectionArgs{
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
			Password: LXD_SERVER_PASSWORD,
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
