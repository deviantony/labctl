package dockerhub

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AccessToken represents a DockerHub access token.
type AccessToken struct {
	Uuid      string    `json:"uuid"`
	Token     string    `json:"token"`
	Label     string    `json:"token_label"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

const (
	accessTokenScopeRead  = "repo:read"
	accessTokenScopeWrite = "repo:write"
)

type (
	// CreateAccessTokenPayload is the payload for the DockerHub API CreateAccessToken endpoint.
	CreateAccessTokenPayload struct {
		Label  string   `json:"token_label"`
		Scopes []string `json:"scopes"`
	}

	// CreateAccessTokenResponse is the response for the DockerHub API ListAccessToken endpoint.
	ListAccessTokensResponse struct {
		Count   int           `json:"count"`
		Results []AccessToken `json:"results"`
	}
)

// CreateAccessToken creates a new access token.
func (c *Client) CreateAccessToken(label string) (AccessToken, error) {
	url := "https://hub.docker.com/v2/access-tokens"

	payload := CreateAccessTokenPayload{
		Label:  label,
		Scopes: []string{accessTokenScopeRead, accessTokenScopeWrite},
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return AccessToken{}, fmt.Errorf("unable to marshal JSON for DockerHub API CreateAccessToken payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return AccessToken{}, fmt.Errorf("unable to create DockerHub API CreateAccessToken request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return AccessToken{}, fmt.Errorf("unable to execute DockerHub API CreateAccessToken request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return AccessToken{}, fmt.Errorf("unexpected DockerHub API CreateAccessToken response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AccessToken{}, fmt.Errorf("unable to read DockerHub API CreateAccessToken response: %w", err)
	}

	var responseData AccessToken
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return AccessToken{}, fmt.Errorf("unable to unmarshal JSON for DockerHub API CreateAccessToken response: %w", err)
	}

	return responseData, nil
}

// DeleteAccessToken deletes an access token based on the specified UUID.
func (c *Client) DeleteAccessToken(uuid string) error {
	url := fmt.Sprintf("https://hub.docker.com/v2/access-tokens/%s", uuid)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("unable to create DockerHub API DeleteAccessToken request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("unable to execute DockerHub API DeleteAccessToken request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected DockerHub API DeleteAccessToken response status: %d", resp.StatusCode)
	}

	return nil
}

// ListAccessTokens lists all access tokens.
func (c *Client) ListAccessTokens() ([]AccessToken, error) {
	url := "https://hub.docker.com/v2/access-tokens"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []AccessToken{}, fmt.Errorf("unable to create DockerHub AccessTokenList request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return []AccessToken{}, fmt.Errorf("unable to execute DockerHub API AccessTokenList request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []AccessToken{}, fmt.Errorf("unexpected DockerHub API AccessTokenList response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []AccessToken{}, fmt.Errorf("unable to read DockerHub API AccessTokenList response: %w", err)
	}

	var responseData ListAccessTokensResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return []AccessToken{}, fmt.Errorf("unable to unmarshal JSON for DockerHub API AccessTokenList response: %w", err)
	}

	return responseData.Results, nil
}
