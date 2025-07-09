package twitch

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
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

// Drop-related methods
func (c *Client) GetDropCampaigns(ctx context.Context) ([]Campaign, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	campaigns, err := gqlClient.GetCampaigns(ctx)
	if err != nil {
		// Check if this is an authentication error
		if c.isAuthError(err) {
			logrus.Info("Token appears invalid, clearing stored token")
			c.clearToken()
			return nil, fmt.Errorf("authentication expired, please re-login")
		}
		return nil, fmt.Errorf("failed to get drop campaigns: %w", err)
	}

	return campaigns, nil
}

// GetCampaignDetails returns detailed information about a specific campaign
func (c *Client) GetCampaignDetails(ctx context.Context, campaignID string) (*Campaign, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	user := c.user
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	if user == nil {
		return nil, fmt.Errorf("user not available")
	}

	campaign, err := gqlClient.GetCampaignDetails(ctx, campaignID, user.Login)
	if err != nil {
		// Check if this is an authentication error
		if c.isAuthError(err) {
			logrus.Info("Token appears invalid, clearing stored token")
			c.clearToken()
			return nil, fmt.Errorf("authentication expired, please re-login")
		}
		return nil, fmt.Errorf("failed to get campaign details: %w", err)
	}

	return campaign, nil
}

// GetCurrentDropProgress returns current drop progress using TDM's DropCurrentSessionContext
func (c *Client) GetCurrentDropProgress(ctx context.Context, channelID string) (*CurrentDropProgress, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	// Use TDM's exact DropCurrentSessionContext operation
	operation, err := GetOperation("CurrentDrop", map[string]interface{}{
		"channelID":    channelID, // channel ID as string
		"channelLogin": "",        // always empty string per TDM
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get CurrentDrop operation: %w", err)
	}

	resp, err := gqlClient.GQLRequest(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute CurrentDrop query: %w", err)
	}

	// Parse response using TDM's exact path
	return c.parseCurrentDropResponse(resp.Data)
}

// parseCurrentDropResponse parses the CurrentDrop response using TDM's exact approach
func (c *Client) parseCurrentDropResponse(data interface{}) (*CurrentDropProgress, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	// Follow TDM's exact path: context["data"]["currentUser"]["dropCurrentSession"]
	currentUser, ok := dataMap["currentUser"]
	if !ok || currentUser == nil {
		logrus.Debugf("No currentUser in CurrentDrop response")
		return nil, fmt.Errorf("no currentUser in response")
	}

	userMap, ok := currentUser.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid currentUser format")
	}

	dropCurrentSession, ok := userMap["dropCurrentSession"]
	if !ok || dropCurrentSession == nil {
		logrus.Debugf("No dropCurrentSession in response - no active drop progress")
		return nil, nil // This is normal if no drop is being tracked
	}

	sessionMap, ok := dropCurrentSession.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid dropCurrentSession format")
	}

	// Extract progress data like TDM does
	progress := &CurrentDropProgress{}

	if currentMinutes, ok := sessionMap["currentMinutesWatched"].(float64); ok {
		progress.CurrentMinutesWatched = int(currentMinutes)
	}

	progress.DropID = getString(sessionMap, "dropID")

	logrus.Infof("=== SUCCESS: Real Progress from DropCurrentSessionContext ===")
	logrus.Infof("Drop ID: %s, Current Minutes: %d", progress.DropID, progress.CurrentMinutesWatched)

	return progress, nil
}

// parseAvailableDropsResponse parses the AvailableDrops response to find progress
func (c *Client) parseAvailableDropsResponse(data interface{}) (*CurrentDropProgress, error) {
	// This function is not used anymore since we switched back to DropCurrentSessionContext
	return nil, fmt.Errorf("AvailableDrops parsing not implemented")
}

func (c *Client) isAuthError(err error) bool {
	// Check if error indicates authentication issues
	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "invalid token") ||
		strings.Contains(errStr, "token validation failed")
}

func (c *Client) clearToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = nil
	c.user = nil
	c.isLoggedIn = false
	c.gqlClient = nil // Clear TDM GraphQL client
	config.DeleteToken()
}

func (c *Client) GetStreamsForGame(ctx context.Context, gameSlug string, limit int) ([]Stream, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	streams, err := gqlClient.GetStreamsForGame(ctx, gameSlug, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get streams for game: %w", err)
	}

	return streams, nil
}

// GetStreamsForGameName gets streams for a game by name, resolving the slug if needed
func (c *Client) GetStreamsForGameName(ctx context.Context, gameName string, limit int) ([]Stream, error) {
	// First try to get the slug from the already resolved game name
	slugInfo, err := c.GetGameSlug(ctx, gameName)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve game slug for '%s': %w", gameName, err)
	}

	// Now get streams using the resolved slug
	return c.GetStreamsForGame(ctx, slugInfo.Slug, limit)
}

// GetGameSlug converts a game name to its Twitch slug and ID
func (c *Client) GetGameSlug(ctx context.Context, gameName string) (*GameSlugInfo, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	slugInfo, err := gqlClient.GetGameSlug(ctx, gameName)
	if err != nil {
		return nil, fmt.Errorf("failed to get game slug: %w", err)
	}

	return slugInfo, nil
}

// StartWatching initiates stream watching like TDM
func (c *Client) StartWatching(ctx context.Context, channelLogin string) (*WatchingSession, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	// Get playback access token first
	token, err := gqlClient.GetPlaybackAccessToken(ctx, channelLogin)
	if err != nil {
		return nil, fmt.Errorf("failed to get playback token: %w", err)
	}

	// Get stream URL
	streamURL, err := gqlClient.GetStreamURL(ctx, channelLogin, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get stream URL: %w", err)
	}

	session := &WatchingSession{
		ChannelLogin: channelLogin,
		StreamURL:    streamURL,
		GQLClient:    gqlClient,
	}

	logrus.Infof("Started watching session for %s", channelLogin)
	return session, nil
}

// SendWatchRequest sends periodic watch request like TDM
func (c *Client) SendWatchRequest(ctx context.Context, session *WatchingSession) error {
	if session == nil || session.GQLClient == nil {
		return fmt.Errorf("invalid watching session")
	}

	return session.GQLClient.SendWatchRequest(ctx, session.StreamURL)
}

func (c *Client) ClaimDrop(ctx context.Context, dropInstanceID string) error {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	if err := gqlClient.ClaimDrop(ctx, dropInstanceID); err != nil {
		return fmt.Errorf("failed to claim drop: %w", err)
	}

	return nil
}

func (c *Client) GetInventory(ctx context.Context) (*Inventory, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	inventory, err := gqlClient.GetInventory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory, nil
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
