package twitch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"twitchdropsfarmer/internal/config"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Client struct {
	authManager *AuthManager
	gqlClient   *GraphQLClient // TDM-style GraphQL client

	// Authentication state
	mu         sync.RWMutex
	token      *oauth2.Token
	user       *User
	isLoggedIn bool

	// TDM-style session data
	sessionID string
	deviceID  string

	// Client configuration
	clientID     string
	clientSecret string
}

// generateNonce generates a random hex string of specified length
func generateNonce(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func NewClient(clientID, clientSecret string) *Client {
	client := &Client{
		authManager:  NewAuthManager(clientID, clientSecret),
		clientID:     clientID,
		clientSecret: clientSecret,
		sessionID:    generateNonce(16), // 16 char hex string like TDM
		deviceID:     generateNonce(32), // 32 char hex string like TDM
	}

	// Try to load existing token
	client.loadStoredToken()

	return client
}

func (c *Client) loadStoredToken() {
	token, err := config.LoadToken()
	if err != nil {
		logrus.Debugf("No stored token found: %v", err)
		return
	}

	// Validate the token before using it (don't check expiry, let Twitch decide)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := c.authManager.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		logrus.Debugf("Stored token invalid: %v", err)
		// Only delete if actually invalid (not just expired according to our local time)
		config.DeleteToken()
		return
	}

	// Token is valid, set it and extend expiry to 1 year like TDM
	token.Expiry = time.Now().Add(365 * 24 * time.Hour) // 1 year

	c.mu.Lock()
	c.token = token
	c.user = user
	c.isLoggedIn = true
	// Initialize TDM-style GraphQL client with token
	c.gqlClient = NewGraphQLClient(token.AccessToken, c.sessionID, c.deviceID)
	c.mu.Unlock()

	// Save the token with extended expiry
	if err := config.SaveToken(token); err != nil {
		logrus.Errorf("Failed to save extended token: %v", err)
	}

	logrus.Infof("Loaded stored authentication for %s", user.DisplayName)
}

// Authentication methods - Device Code Flow (like TDM)
func (c *Client) StartDeviceFlow(ctx context.Context) (*DeviceCodeResponse, error) {
	return c.authManager.GenerateDeviceCode(ctx)
}

func (c *Client) PollForToken(ctx context.Context, deviceCode string, interval int) error {
	token, err := c.authManager.PollForToken(ctx, deviceCode, interval)
	if err != nil {
		return fmt.Errorf("failed to poll for token: %w", err)
	}

	user, err := c.authManager.ValidateToken(ctx, token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to validate token: %w", err)
	}

	// Set token expiry to 1 year (like TDM)
	token.Expiry = time.Now().Add(365 * 24 * time.Hour)

	c.mu.Lock()
	c.token = token
	c.user = user
	c.isLoggedIn = true
	// Initialize TDM-style GraphQL client with token
	c.gqlClient = NewGraphQLClient(token.AccessToken, c.sessionID, c.deviceID)
	c.mu.Unlock()

	// Save token to persistent storage
	if err := config.SaveToken(token); err != nil {
		logrus.Errorf("Failed to save token: %v", err)
	} else {
		logrus.Debug("Token saved to persistent storage")
	}

	logrus.Infof("Successfully authenticated as %s", user.DisplayName)
	return nil
}

func (c *Client) IsLoggedIn() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isLoggedIn
}

func (c *Client) GetUser() *User {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.user == nil {
		return nil
	}
	userCopy := *c.user
	return &userCopy
}

func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != nil {
		if err := c.authManager.RevokeToken(ctx, c.token.AccessToken); err != nil {
			logrus.Warnf("Failed to revoke token: %v", err)
		}
	}

	// Delete stored token
	if err := config.DeleteToken(); err != nil {
		logrus.Warnf("Failed to delete stored token: %v", err)
	}

	c.token = nil
	c.user = nil
	c.isLoggedIn = false

	logrus.Info("Successfully logged out")
	return nil
}

func (c *Client) refreshTokenIfNeeded(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token == nil {
		return fmt.Errorf("no token available")
	}

	// Don't check local expiry - let Twitch tell us if the token is invalid
	// This is more like TDM behavior where tokens work until Twitch says they don't
	return nil
}

func (c *Client) getAccessToken(ctx context.Context) (string, error) {
	if err := c.refreshTokenIfNeeded(ctx); err != nil {
		return "", err
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.token == nil {
		return "", fmt.Errorf("no access token available")
	}

	return c.token.AccessToken, nil
}

// Utility methods
func (c *Client) GetClientID() string {
	return c.clientID
}

func (c *Client) SetToken(token *oauth2.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
}

func (c *Client) GetToken() *oauth2.Token {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.token == nil {
		return nil
	}
	// Return a copy to prevent external modification
	tokenCopy := *c.token
	return &tokenCopy
}
