package twitch

import (
	"context"
	"fmt"
	"strings"

	"twitchdropsfarmer/internal/config"

	"github.com/sirupsen/logrus"
)

// GraphQL-specific client methods

// getGQLClient safely retrieves the GraphQL client or returns an error if not authenticated
func (c *Client) getGQLClient() (*GraphQLClient, error) {
	c.mu.RLock()
	gqlClient := c.gqlClient
	c.mu.RUnlock()

	if gqlClient == nil {
		return nil, fmt.Errorf("not authenticated - GraphQL client not initialized")
	}

	return gqlClient, nil
}

// GetDropCampaigns retrieves all available drop campaigns
func (c *Client) GetDropCampaigns(ctx context.Context) ([]Campaign, error) {
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
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
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	user := c.user
	c.mu.RUnlock()

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
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
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

// GetStreamsForGame retrieves streams for a specific game slug
func (c *Client) GetStreamsForGame(ctx context.Context, gameSlug string, limit int) ([]Stream, error) {
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
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
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
	}

	slugInfo, err := gqlClient.GetGameSlug(ctx, gameName)
	if err != nil {
		return nil, fmt.Errorf("failed to get game slug: %w", err)
	}

	return slugInfo, nil
}

// StartWatching initiates stream watching like TDM
func (c *Client) StartWatching(ctx context.Context, channelLogin string) (*WatchingSession, error) {
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
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

// ClaimDrop claims a completed drop
func (c *Client) ClaimDrop(ctx context.Context, dropInstanceID string) error {
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return err
	}

	if err := gqlClient.ClaimDrop(ctx, dropInstanceID); err != nil {
		return fmt.Errorf("failed to claim drop: %w", err)
	}

	return nil
}

// GetInventory retrieves user's drop inventory
func (c *Client) GetInventory(ctx context.Context) (*Inventory, error) {
	gqlClient, err := c.getGQLClient()
	if err != nil {
		return nil, err
	}

	inventory, err := gqlClient.GetInventory(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inventory, nil
}

// Helper methods for GraphQL operations

// isAuthError checks if error indicates authentication issues
func (c *Client) isAuthError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "invalid token") ||
		strings.Contains(errStr, "token validation failed")
}

// clearToken clears the stored token and authentication state
func (c *Client) clearToken() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.token = nil
	c.user = nil
	c.isLoggedIn = false
	c.gqlClient = nil // Clear TDM GraphQL client
	config.DeleteToken()
}
