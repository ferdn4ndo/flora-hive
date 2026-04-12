package userver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"flora-hive/lib"
)

// Client calls uServer-Auth HTTP APIs.
type Client struct {
	http    *http.Client
	baseURL string
	sysName string
	sysTok  string
}

// NewClient builds a Client; baseURL should have no trailing slash.
func NewClient(env lib.Env) *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: strings.TrimSuffix(strings.TrimSpace(env.UserverAuthHost), "/"),
		sysName: strings.TrimSpace(env.UserverAuthSystemName),
		sysTok:  strings.TrimSpace(env.UserverAuthSystemToken),
	}
}

// MeResponse mirrors /auth/me JSON.
type MeResponse struct {
	UUID           string `json:"uuid"`
	SystemName     string `json:"system_name"`
	Username       string `json:"username"`
	RegisteredAt   string `json:"registered_at"`
	LastActivityAt string `json:"last_activity_at"`
	IsAdmin        bool   `json:"is_admin"`
	Token          struct {
		IssuedAt  string `json:"issued_at"`
		ExpiresAt string `json:"expires_at"`
	} `json:"token"`
}

// LoginResponse mirrors login/refresh JSON.
type LoginResponse struct {
	AccessToken     string `json:"access_token"`
	AccessTokenExp  string `json:"access_token_exp"`
	RefreshToken    string `json:"refresh_token"`
	RefreshTokenExp string `json:"refresh_token_exp"`
}

// RegisterResponse mirrors register JSON.
type RegisterResponse struct {
	Username   string        `json:"username"`
	SystemName string        `json:"system_name"`
	IsAdmin    bool          `json:"is_admin"`
	Auth       LoginResponse `json:"auth"`
}

type apiErr struct {
	Message string `json:"message"`
}

// doRequest executes the request and returns the full response body (connection closed).
func (c *Client) doRequest(req *http.Request) (status int, body []byte, err error) {
	res, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return res.StatusCode, nil, err
	}
	return res.StatusCode, b, nil
}

func parseAPIErrMessage(body []byte, statusCode int) string {
	var e apiErr
	if err := json.Unmarshal(bytes.TrimSpace(body), &e); err == nil && e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("request failed (%d)", statusCode)
}

// Login calls POST /auth/login.
func (c *Client) Login(username, password string) (*LoginResponse, int, string, error) {
	body, err := json.Marshal(map[string]string{
		"username":     username,
		"password":     password,
		"system_name":  c.sysName,
		"system_token": c.sysTok,
	})
	if err != nil {
		return nil, 0, "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	status, respBody, err := c.doRequest(req)
	if err != nil {
		return nil, 0, "", err
	}
	if status < 200 || status >= 300 {
		return nil, status, parseAPIErrMessage(respBody, status), nil
	}
	var out LoginResponse
	if len(bytes.TrimSpace(respBody)) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, status, "", err
		}
	}
	return &out, status, "", nil
}

// Register calls POST /auth/register.
func (c *Client) Register(username, password string, isAdmin *bool) (*RegisterResponse, int, string, error) {
	payload := map[string]interface{}{
		"username":     username,
		"password":     password,
		"system_name":  c.sysName,
		"system_token": c.sysTok,
	}
	if isAdmin != nil {
		payload["is_admin"] = *isAdmin
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, "", err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/register", bytes.NewReader(body))
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	status, respBody, err := c.doRequest(req)
	if err != nil {
		return nil, 0, "", err
	}
	if status < 200 || status >= 300 {
		return nil, status, parseAPIErrMessage(respBody, status), nil
	}
	var out RegisterResponse
	if len(bytes.TrimSpace(respBody)) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, status, "", err
		}
	}
	return &out, status, "", nil
}

// Refresh calls POST /auth/refresh with Bearer refresh token.
func (c *Client) Refresh(refreshToken string) (*LoginResponse, int, string, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/refresh", nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+refreshToken)
	req.Header.Set("Content-Type", "application/json")
	status, respBody, err := c.doRequest(req)
	if err != nil {
		return nil, 0, "", err
	}
	if status < 200 || status >= 300 {
		return nil, status, parseAPIErrMessage(respBody, status), nil
	}
	var out LoginResponse
	if len(bytes.TrimSpace(respBody)) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, status, "", err
		}
	}
	return &out, status, "", nil
}

// Logout calls POST /auth/logout.
func (c *Client) Logout(accessToken string) (int, error) {
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/auth/logout", nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	status, _, err := c.doRequest(req)
	return status, err
}

// Me calls GET /auth/me.
func (c *Client) Me(accessToken string) (*MeResponse, int, string, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/auth/me", nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	status, respBody, err := c.doRequest(req)
	if err != nil {
		return nil, 0, "", err
	}
	if status < 200 || status >= 300 {
		return nil, status, parseAPIErrMessage(respBody, status), nil
	}
	var out MeResponse
	if len(bytes.TrimSpace(respBody)) > 0 {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return nil, status, "", err
		}
	}
	if strings.TrimSpace(out.UUID) == "" {
		return nil, 502, "Invalid me response", nil
	}
	return &out, status, "", nil
}

// ChangePassword calls PATCH /auth/me/password.
func (c *Client) ChangePassword(accessToken string, currentPassword, newPassword string) (int, string, error) {
	body, err := json.Marshal(map[string]string{
		"current_password": currentPassword,
		"new_password":     newPassword,
	})
	if err != nil {
		return 0, "", err
	}
	req, err := http.NewRequest(http.MethodPatch, c.baseURL+"/auth/me/password", bytes.NewReader(body))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	status, respBody, err := c.doRequest(req)
	if err != nil {
		return 0, "", err
	}
	if status < 200 || status >= 300 {
		return status, parseAPIErrMessage(respBody, status), nil
	}
	return status, "", nil
}
