package dockerhub

import (
	"fmt"
	"time"

	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

// DockerHubClient is a high-level wrapper around the DockerHub API client.
type DockerHubClient struct {
	Username string
	Password string

	timeout time.Duration
	code    string
	logger  *zap.SugaredLogger
}

// NewDockerHubClient creates a new DockerHub API client.
func NewDockerHubClient(cfg config.DockerHubConfig, logger *zap.SugaredLogger, twoFACode string) *DockerHubClient {
	return &DockerHubClient{
		Username: cfg.Username,
		Password: cfg.Password,
		timeout:  cfg.Timeout,
		logger:   logger,
		code:     twoFACode,
	}
}

// ListAccessTokens lists all access tokens.
func (c *DockerHubClient) ListAccessTokens() ([]AccessToken, error) {
	client := NewClient(c.timeout, c.logger)

	authToken, err := client.LoginWith2FA(c.Username, c.Password, c.code)
	if err != nil {
		return nil, err
	}

	return client.ListAccessTokens(authToken)
}

// CreateAccessToken creates a new access token.
// If label is not specified, it will default to "labctl-DATETIME".
func (c *DockerHubClient) CreateAccessToken(label string) (string, error) {
	if label == "" {
		label = fmt.Sprintf("labctl-%s", time.Now().Format(time.RFC3339))
	}

	client := NewClient(c.timeout, c.logger)

	authToken, err := client.LoginWith2FA(c.Username, c.Password, c.code)
	if err != nil {
		return "", err
	}

	return client.CreateAccessToken(authToken, label)
}

// DeleteAccessToken deletes an access token.
func (c *DockerHubClient) DeleteAccessToken(uuid string) error {
	client := NewClient(c.timeout, c.logger)

	authToken, err := client.LoginWith2FA(c.Username, c.Password, c.code)
	if err != nil {
		return err
	}

	return client.DeleteAccessToken(authToken, uuid)
}
