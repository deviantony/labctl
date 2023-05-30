package dockerhub

import (
	"fmt"
	"time"

	"github.com/deviantony/labctl/internal/config"
	"go.uber.org/zap"
)

// DockerHubClient is a high-level wrapper around the DockerHub API client.
type DockerHubClient struct {
	username  string
	password  string
	twoFAcode string
	timeout   time.Duration
	logger    *zap.SugaredLogger
	authToken string
}

// NewDockerHubClient creates a new DockerHub API client.
func NewDockerHubClient(cfg config.DockerHubConfig, logger *zap.SugaredLogger, twoFACode string) *DockerHubClient {
	return &DockerHubClient{
		username:  cfg.Username,
		password:  cfg.Password,
		timeout:   cfg.Timeout,
		logger:    logger,
		twoFAcode: twoFACode,
	}
}

// CreateAccessToken creates a new access token.
// If label is not specified, it will default to "labctl-DATETIME".
func (c *DockerHubClient) CreateAccessToken(label string) (AccessToken, error) {
	if label == "" {
		label = fmt.Sprintf("labctl-%s", time.Now().Format(time.RFC3339))
	}

	err := c.login()
	if err != nil {
		return AccessToken{}, err
	}

	client := NewClient(c.authToken, c.timeout, c.logger)
	return client.CreateAccessToken(label)
}

// DeleteAccessToken deletes an access token.
func (c *DockerHubClient) DeleteAccessToken(uuid string) error {
	err := c.login()
	if err != nil {
		return err
	}

	client := NewClient(c.authToken, c.timeout, c.logger)
	return client.DeleteAccessToken(uuid)
}

// ListAccessTokens lists all access tokens.
func (c *DockerHubClient) ListAccessTokens() ([]AccessToken, error) {
	err := c.login()
	if err != nil {
		return nil, err
	}

	client := NewClient(c.authToken, c.timeout, c.logger)
	return client.ListAccessTokens()
}

func (c *DockerHubClient) login() error {
	if c.authToken != "" {
		return nil
	}

	client := NewClient(c.authToken, c.timeout, c.logger)

	authToken, err := client.LoginWith2FA(c.username, c.password, c.twoFAcode)
	if err != nil {
		return err
	}

	c.authToken = authToken
	return nil
}
