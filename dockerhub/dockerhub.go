package dockerhub

import (
	"time"

	"github.com/deviantony/labctl/config"
	"go.uber.org/zap"
)

type DockerHubClient struct {
	Username string
	Password string

	timeout time.Duration
	code    string
	logger  *zap.SugaredLogger
}

func NewDockerHubClient(cfg config.DockerHubConfig, logger *zap.SugaredLogger, twoFACode string) *DockerHubClient {
	return &DockerHubClient{
		Username: cfg.Username,
		Password: cfg.Password,
		timeout:  cfg.Timeout,
		logger:   logger,
		code:     twoFACode,
	}
}

func (c *DockerHubClient) ListAccessTokens() ([]AccessToken, error) {
	client := NewClient(c.timeout, c.logger)

	accessToken, err := client.LoginWith2FA(c.Username, c.Password, c.code)
	if err != nil {
		return nil, err
	}

	return client.ListAccessTokens(accessToken)
}
