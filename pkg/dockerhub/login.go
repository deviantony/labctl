package dockerhub

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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
)

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
		return "", fmt.Errorf("unable to marshal JSON for DockerHub API Login payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("unable to create DockerHub API Login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to execute DockerHub API Login request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		return "", fmt.Errorf("unexpected DockerHub API Login response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read DockerHub API Login response: %w", err)
	}

	var responseData LoginResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal JSON for DockerHub API Login response: %w", err)
	}

	if responseData.TwoFAToken == "" {
		return "", errors.New("2FA token not found in DockerHub API Login response")
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
		return "", fmt.Errorf("unable to marshal JSON for DockerHub API 2FALogin payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return "", fmt.Errorf("unable to create DockerHub API 2FALogin request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: c.timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to execute DockerHub API 2FALogin request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected DockerHub API 2FALogin response status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read DockerHub API 2FALogin response: %w", err)
	}

	var responseData TwoFALoginResponse
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshal JSON for DockerHub API 2FALogin response: %w", err)
	}

	return responseData.Token, nil
}
