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

type Client struct {
	timeout time.Duration
	logger  *zap.SugaredLogger
}

func NewClient(timeout time.Duration, logger *zap.SugaredLogger) *Client {
	return &Client{
		timeout: timeout,
		logger:  logger,
	}
}

type AccessToken struct {
	Uuid      string    `json:"uuid"`
	Token     string    `json:"token"`
	Label     string    `json:"token_label"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

type (
	LoginPayload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	LoginResponse struct {
		TwoFAToken string `json:"login_2fa_token"`
	}

	TwoFALoginPayload struct {
		TwoFAToken string `json:"login_2fa_token"`
		Code       string `json:"code"`
	}

	TwoFALoginResponse struct {
		Token string `json:"token"`
	}

	CreateAccessTokenPayload struct {
		Label  string   `json:"token_label"`
		Scopes []string `json:"scopes"`
	}

	ListAccessTokensResponse struct {
		Count   int           `json:"count"`
		Results []AccessToken `json:"results"`
	}
)

// ListAccessTokens lists all access tokens for the authenticated user.
func (c *Client) ListAccessTokens(token string) ([]AccessToken, error) {
	url := "https://hub.docker.com/v2/access-tokens"

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c.logger.Errorw("Unable to create DockerHub AccessTokenList request",
			"error", err,
			"url", url,
		)

		return []AccessToken{}, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

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
