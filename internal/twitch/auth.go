package twitch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	DeviceCodeURL = "https://id.twitch.tv/oauth2/device"
	TokenURL      = "https://id.twitch.tv/oauth2/token"
	ValidateURL   = "https://id.twitch.tv/oauth2/validate"
)

var (
	// TDM uses NO scopes - empty string (matching exactly)
	RequiredScopes = []string{
		// Intentionally empty - TDM requests no scopes for clean auth page
	}
)

type AuthManager struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
}

type DeviceTokenPollResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewAuthManager(clientID, clientSecret string) *AuthManager {
	return &AuthManager{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GenerateDeviceCode initiates the device code flow like TDM
func (a *AuthManager) GenerateDeviceCode(ctx context.Context) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", a.clientID)
	data.Set("scopes", strings.Join(RequiredScopes, " "))

	req, err := http.NewRequestWithContext(ctx, "POST", DeviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create device code request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device code request failed with status: %d", resp.StatusCode)
	}

	var deviceResp DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&deviceResp); err != nil {
		return nil, fmt.Errorf("failed to decode device code response: %w", err)
	}

	return &deviceResp, nil
}

// PollForToken polls for the access token after user activates device
func (a *AuthManager) PollForToken(ctx context.Context, deviceCode string, interval int) (*oauth2.Token, error) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Set a timeout for the polling (15 minutes like TDM)
	timeout := time.NewTimer(15 * time.Minute)
	defer timeout.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout.C:
			return nil, fmt.Errorf("device code polling timed out")
		case <-ticker.C:
			token, err := a.checkDeviceCodeStatus(ctx, deviceCode)
			if err != nil {
				logrus.Debugf("Device code polling error: %v", err)
				continue
			}
			if token != nil {
				return token, nil
			}
		}
	}
}

func (a *AuthManager) checkDeviceCodeStatus(ctx context.Context, deviceCode string) (*oauth2.Token, error) {
	data := url.Values{}
	data.Set("client_id", a.clientID)
	// Note: client_secret not required for Twitch Android app device flow
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp DeviceTokenPollResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	// Handle specific errors
	switch tokenResp.Error {
	case "authorization_pending":
		// User hasn't authorized yet, continue polling
		return nil, nil
	case "slow_down":
		// Polling too fast, wait longer
		return nil, fmt.Errorf("polling too fast")
	case "expired_token":
		return nil, fmt.Errorf("device code expired")
	case "access_denied":
		return nil, fmt.Errorf("user denied authorization")
	case "":
		// No error, we have a token
		break
	default:
		return nil, fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDescription)
	}

	if tokenResp.AccessToken == "" {
		return nil, nil // Still waiting
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return token, nil
}

func (a *AuthManager) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	data := url.Values{}
	data.Set("client_id", a.clientID)
	// Note: client_secret not required for Twitch Android app
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	req, err := http.NewRequestWithContext(ctx, "POST", TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode refresh response: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		Expiry:       time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return token, nil
}

func (a *AuthManager) ValidateToken(ctx context.Context, accessToken string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", ValidateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create validate request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token validation failed with status: %d", resp.StatusCode)
	}

	var validateResp struct {
		ClientID  string   `json:"client_id"`
		Login     string   `json:"login"`
		Scopes    []string `json:"scopes"`
		UserID    string   `json:"user_id"`
		ExpiresIn int      `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&validateResp); err != nil {
		return nil, fmt.Errorf("failed to decode validate response: %w", err)
	}

	// Get full user details
	return a.getUserDetails(ctx, accessToken, validateResp.UserID)
}

func (a *AuthManager) getUserDetails(ctx context.Context, accessToken, userID string) (*User, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.twitch.tv/helix/users?id="+userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Client-Id", a.clientID)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user details request failed with status: %d", resp.StatusCode)
	}

	var userResp struct {
		Data []User `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %w", err)
	}

	if len(userResp.Data) == 0 {
		return nil, fmt.Errorf("no user data returned")
	}

	return &userResp.Data[0], nil
}

func (a *AuthManager) RevokeToken(ctx context.Context, accessToken string) error {
	data := url.Values{}
	data.Set("client_id", a.clientID)
	data.Set("token", accessToken)

	req, err := http.NewRequestWithContext(ctx, "POST", "https://id.twitch.tv/oauth2/revoke", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create revoke request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logrus.Warnf("Token revocation returned status: %d", resp.StatusCode)
	}

	return nil
}
