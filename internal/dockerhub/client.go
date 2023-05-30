package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Client is a DockerHub API client.
type Client struct {
	authToken string
	timeout   time.Duration
	logger    *zap.SugaredLogger
}

// NewClient creates a new DockerHub API client.
func NewClient(authToken string, timeout time.Duration, logger *zap.SugaredLogger) *Client {
	return &Client{
		authToken: authToken,
		timeout:   timeout,
		logger:    logger,
	}
}

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
	// LoginPayload is the payload for the DockerHub API Login endpoint.
	LoginPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	// LoginResponse is the response for the DockerHub API Login endpoint.
	LoginResponse struct {
		TwoFAToken string `json:"login_2fa_token"`
	}

	// TwoFALoginPayload is the payload for the DockerHub API 2FA Login endpoint.
	TwoFALoginPayload struct {
		TwoFAToken string `json:"login_2fa_token"`
		Code       string `json:"code"`
	}

	// TwoFALoginResponse is the response for the DockerHub API 2FA Login endpoint.
	TwoFALoginResponse struct {
		Token string `json:"token"`
	}

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
		c.logger.Errorw("Unable to marshal JSON for DockerHub API CreateAccessToken payload",
			"error", err,
			"url", url,
		)

		return AccessToken{}, err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub API CreateAccessToken request",
			"error", err,
			"url", url,
		)

		return AccessToken{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("Unable to execute DockerHub API CreateAccessToken request",
			"error", err,
			"url", url,
		)

		return AccessToken{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		c.logger.Errorw("Unexpected DockerHub API CreateAccessToken response status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"url", url,
		)

		return AccessToken{}, errors.New("unexpected response status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorw("Unable to read DockerHub API CreateAccessToken response",
			"error", err,
			"url", url,
		)

		return AccessToken{}, err
	}

	var responseData AccessToken
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		c.logger.Errorw("Unable to unmarshal JSON for DockerHub API CreateAccessToken response",
			"error", err,
			"url", url,
		)

		return AccessToken{}, err
	}

	return responseData, nil
}

// DeleteAccessToken deletes an access token based on the specified UUID.
func (c *Client) DeleteAccessToken(uuid string) error {
	url := fmt.Sprintf("https://hub.docker.com/v2/access-tokens/%s", uuid)

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub API DeleteAccessToken request",
			"error", err,
			"url", url,
		)

		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("Unable to execute DockerHub API DeleteAccessToken request",
			"error", err,
			"url", url,
		)

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		c.logger.Errorw("Unexpected DockerHub API DeleteAccessToken response status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"url", url,
		)

		return errors.New("unexpected response status")
	}

	return nil
}

// ListAccessTokens lists all access tokens.
func (c *Client) ListAccessTokens() ([]AccessToken, error) {
	url := "https://hub.docker.com/v2/access-tokens"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub AccessTokenList request",
			"error", err,
			"url", url,
		)

		return []AccessToken{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.authToken))

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("Unable to execute DockerHub API AccessTokenList request",
			"error", err,
			"url", url,
		)

		return []AccessToken{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorw("Unexpected DockerHub API AccessTokenList response status",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"url", url,
		)

		return []AccessToken{}, errors.New("unexpected response status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorw("Unable to read DockerHub API AccessTokenList response",
			"error", err,
			"url", url,
		)

		return []AccessToken{}, err
	}

	var responseData ListAccessTokensResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		c.logger.Errorw("Unable to unmarshal JSON for DockerHub API AccessTokenList response",
			"error", err,
			"url", url,
		)

		return []AccessToken{}, err
	}

	return responseData.Results, nil
}

// Login logs in to DockerHub using the provided credentials and a 2FA code.
// This methods requires 2FA to be enabled on the account.
func (c *Client) LoginWith2FA(username, password, code string) (string, error) {
	url := "https://hub.docker.com/v2/users/login/"

	payload := LoginPayload{
		Username: username,
		Password: password,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.logger.Errorw("Unable to marshal JSON for DockerHub API Login payload",
			"error", err,
			"url", url,
		)

		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub API Login request",
			"error", err,
			"url", url,
		)

		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("Unable to execute DockerHub API Login request",
			"error", err,
			"url", url,
		)

		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		c.logger.Errorw("Unexpected DockerHub API Login response status. Make sure 2FA is enabled.",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"url", url,
			"username", username,
		)

		return "", errors.New("unexpected response status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorw("Unable to read DockerHub API Login response",
			"error", err,
			"url", url,
		)

		return "", err
	}

	var responseData LoginResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		c.logger.Errorw("Unable to unmarshal JSON for DockerHub API Login response",
			"error", err,
			"url", url,
		)

		return "", err
	}

	if responseData.TwoFAToken == "" {
		c.logger.Errorw("2FA token not found in DockerHub API Login response",
			"url", url,
		)

		return "", errors.New("2FA token not found in response")
	}

	return c.twoFALogin(responseData.TwoFAToken, code)
}

func (c *Client) twoFALogin(twoFAToken, twoFACode string) (string, error) {
	url := "https://hub.docker.com/v2/users/2fa-login"

	payload := TwoFALoginPayload{
		TwoFAToken: twoFAToken,
		Code:       twoFACode,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		c.logger.Errorw("Unable to marshal JSON for DockerHub API 2FALogin payload",
			"error", err,
			"url", url,
		)

		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub API 2FALogin request",
			"error", err,
			"url", url,
		)

		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Errorw("Unable to execute DockerHub API 2FALogin request",
			"error", err,
			"url", url,
		)

		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Errorw("Unexpected DockerHub API 2FALogin response status.",
			"status_code", resp.StatusCode,
			"status", resp.Status,
			"url", url,
			"code", twoFACode,
		)

		return "", errors.New("unexpected response status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorw("Unable to read DockerHub API 2FALogin response",
			"error", err,
			"url", url,
		)

		return "", err
	}

	var responseData TwoFALoginResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		c.logger.Errorw("Unable to unmarshal JSON for DockerHub API 2FALogin response",
			"error", err,
			"url", url,
		)

		return "", err
	}

	return responseData.Token, nil
}
