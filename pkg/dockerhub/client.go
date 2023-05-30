package dockerhub

import (
	"time"
)

// Client is a DockerHub API client.
type Client struct {
	authToken string
	timeout   time.Duration
}

// NewClient creates a new DockerHub API client.
func NewClient(authToken string, timeout time.Duration) *Client {
	return &Client{
		authToken: authToken,
		timeout:   timeout,
	}
}
