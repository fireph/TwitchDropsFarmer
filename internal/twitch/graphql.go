package twitch

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	GraphQLEndpoint = "https://gql.twitch.tv/gql"
)

// GraphQLClient handles GraphQL requests to Twitch, exactly like TDM
type GraphQLClient struct {
	httpClient  *http.Client
	clientInfo  *ClientInfo
	accessToken string
	sessionID   string
	deviceID    string
}

// ClientInfo matches TDM's ClientType.ANDROID_APP
type ClientInfo struct {
	ClientURL string
	ClientID  string
	UserAgent string
}

// NewGraphQLClient creates a new GraphQL client with TDM's exact configuration
func NewGraphQLClient(accessToken, sessionID, deviceID string) *GraphQLClient {
	// Use TDM's exact Android app client info
	clientInfo := &ClientInfo{
		ClientURL: "https://www.twitch.tv",
		ClientID:  "kd1unb4b3q4t58fwlpcbzcbnm76a8fp",
		UserAgent: "Dalvik/2.1.0 (Linux; U; Android 7.1.2; SM-G977N Build/LMY48Z) tv.twitch.android.app/16.8.1/1608010",
	}

	return &GraphQLClient{
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		clientInfo:  clientInfo,
		accessToken: accessToken,
		sessionID:   sessionID,
		deviceID:    deviceID,
	}
}

// Headers creates request headers exactly like TDM's _AuthState.headers method
func (g *GraphQLClient) Headers(gql bool) map[string]string {
	headers := map[string]string{
		"Accept":          "*/*",
		"Accept-Encoding": "gzip",
		"Accept-Language": "en-US",
		"Pragma":          "no-cache",
		"Cache-Control":   "no-cache",
		"Client-Id":       g.clientInfo.ClientID,
		"User-Agent":      g.clientInfo.UserAgent,
	}

	if g.sessionID != "" {
		headers["Client-Session-Id"] = g.sessionID
	}

	if g.deviceID != "" {
		headers["X-Device-Id"] = g.deviceID
	}

	if gql {
		headers["Origin"] = g.clientInfo.ClientURL
		headers["Referer"] = g.clientInfo.ClientURL
		headers["Authorization"] = fmt.Sprintf("OAuth %s", g.accessToken)
	}

	return headers
}

// GQLRequest executes GraphQL requests exactly like TDM's gql_request method
func (g *GraphQLClient) GQLRequest(ctx context.Context, operation *GQLOperation) (*GraphQLResponse, error) {
	// No GraphQL logging

	// Convert operation to JSON
	jsonBody, err := operation.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", GraphQLEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers exactly like TDM
	headers := g.Headers(true)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// No GraphQL headers logging

	// Execute request
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// No GraphQL status logging

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GraphQL request failed with status: %d", resp.StatusCode)
	}

	// Handle gzip compression
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// Parse response
	var gqlResp GraphQLResponse
	if err := json.NewDecoder(reader).Decode(&gqlResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Handle GraphQL errors like TDM
	if len(gqlResp.Errors) > 0 {
		for _, err := range gqlResp.Errors {
			if err.Message == "service error" || err.Message == "PersistedQueryNotFound" {
				logrus.Errorf("Retrying a %s for %s", err.Message, operation.OperationName)
				// TDM would retry here, but for now we'll just log and continue
			}
		}
		return &gqlResp, fmt.Errorf("GraphQL errors: %v", gqlResp.Errors)
	}

	return &gqlResp, nil
}

func (g *GraphQLClient) executeOperation(ctx context.Context, opType OperationType, variables map[string]interface{}) (*GraphQLResponse, error) {
	operation, err := GetOperation(opType, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s operation: %w", opType.String(), err)
	}

	resp, err := g.GQLRequest(ctx, operation)
	if err != nil {
		return nil, fmt.Errorf("failed to execute %s query: %w", opType.String(), err)
	}

	return resp, nil
}

// GetCampaigns fetches drop campaigns using TDM's exact approach
func (g *GraphQLClient) GetCampaigns(ctx context.Context) ([]Campaign, error) {
	resp, err := g.executeOperation(ctx, OpCampaigns, nil)
	if err != nil {
		return nil, err
	}

	// Parse campaigns from response
	campaigns, err := g.parseCampaignsResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse campaigns: %w", err)
	}

	return campaigns, nil
}

// GetInventory fetches drop inventory using TDM's exact approach
func (g *GraphQLClient) GetInventory(ctx context.Context) (*InventoryGQL, error) {
	resp, err := g.executeOperation(ctx, OpInventory, nil)
	if err != nil {
		return nil, err
	}

	// Parse inventory from response
	inventory, err := g.parseInventoryResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inventory: %w", err)
	}

	return inventory, nil
}

// GetStreamsForGame fetches live streams for a specific game using TDM's approach
func (g *GraphQLClient) GetStreamsForGame(ctx context.Context, gameSlug string, limit int) ([]Stream, error) {
	resp, err := g.executeOperation(ctx, OpGameDirectory, map[string]interface{}{
		"slug":  gameSlug,
		"limit": limit,
	})
	if err != nil {
		return nil, err
	}

	// Parse streams from response
	streams, err := g.parseStreamsResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse streams: %w", err)
	}

	logrus.Debugf("Found %d streams for game slug '%s'", len(streams), gameSlug)
	return streams, nil
}

// ClaimDrop claims a completed drop using TDM's exact approach
func (g *GraphQLClient) ClaimDrop(ctx context.Context, dropInstanceID string) error {
	resp, err := g.executeOperation(ctx, OpClaimDrop, map[string]interface{}{
		"input": map[string]interface{}{
			"dropInstanceID": dropInstanceID,
		},
	})
	if err != nil {
		return err
	}

	// Check for successful claim (simplified for now)
	logrus.Debugf("Drop claim response: %+v", resp.Data)

	return nil
}

// GetGameSlug converts a game name to its Twitch slug using DirectoryGameRedirect
func (g *GraphQLClient) GetGameSlug(ctx context.Context, gameName string) (*GameSlugInfo, error) {
	resp, err := g.executeOperation(ctx, OpSlugRedirect, map[string]interface{}{
		"name": gameName,
	})
	if err != nil {
		return nil, err
	}

	// Parse slug info from response
	slugInfo, err := g.parseSlugRedirectResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse slug redirect: %w", err)
	}

	return slugInfo, nil
}

// parseSlugRedirectResponse parses the slug redirect response
func (g *GraphQLClient) parseSlugRedirectResponse(data interface{}) (*GameSlugInfo, error) {
	// Debug: Log the full response structure
	responseJSON, _ := json.MarshalIndent(data, "", "  ")
	logrus.Debugf("SlugRedirect response structure:\n%s", string(responseJSON))

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	// Debug: Log available keys
	logrus.Debugf("Available keys in response: %+v", getKeys(dataMap))

	gameDirectory, ok := dataMap["gameDirectory"]
	if !ok || gameDirectory == nil {
		// Let's try different possible field names
		if game, ok := dataMap["game"]; ok {
			gameDirectory = game
		} else if directoryGame, ok := dataMap["directoryGame"]; ok {
			gameDirectory = directoryGame
		} else {
			return nil, fmt.Errorf("no gameDirectory in response (available keys: %+v)", getKeys(dataMap))
		}
	}

	gameDirectoryMap, ok := gameDirectory.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid gameDirectory format")
	}

	// Debug: Log available keys in gameDirectory
	logrus.Debugf("Available keys in gameDirectory: %+v", getKeys(gameDirectoryMap))

	// Extract slug and ID from the response
	slug, ok := gameDirectoryMap["slug"].(string)
	if !ok {
		return nil, fmt.Errorf("no slug in gameDirectory (available keys: %+v)", getKeys(gameDirectoryMap))
	}

	id, ok := gameDirectoryMap["id"].(string)
	if !ok {
		return nil, fmt.Errorf("no id in gameDirectory (available keys: %+v)", getKeys(gameDirectoryMap))
	}

	return &GameSlugInfo{
		ID:   id,
		Slug: slug,
	}, nil
}

// GetPlaybackAccessToken gets stream access token for watching (like TDM)
func (g *GraphQLClient) GetPlaybackAccessToken(ctx context.Context, channelLogin string) (*PlaybackAccessToken, error) {
	resp, err := g.executeOperation(ctx, OpPlaybackAccessToken, map[string]interface{}{
		"login": channelLogin,
	})
	if err != nil {
		return nil, err
	}

	// Parse token from response
	token, err := g.parsePlaybackTokenResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse playback token: %w", err)
	}

	return token, nil
}

// parsePlaybackTokenResponse parses the playback access token response
func (g *GraphQLClient) parsePlaybackTokenResponse(data interface{}) (*PlaybackAccessToken, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	streamPlaybackAccessToken, ok := dataMap["streamPlaybackAccessToken"]
	if !ok || streamPlaybackAccessToken == nil {
		return nil, fmt.Errorf("no streamPlaybackAccessToken in response")
	}

	tokenMap, ok := streamPlaybackAccessToken.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid streamPlaybackAccessToken format")
	}

	token := &PlaybackAccessToken{
		Value:     getString(tokenMap, "value"),
		Signature: getString(tokenMap, "signature"),
	}

	return token, nil
}

// GetStreamURL gets the HLS stream URL for watching (like TDM)
func (g *GraphQLClient) GetStreamURL(ctx context.Context, channelLogin string, token *PlaybackAccessToken) (string, error) {
	// Build the HLS URL like TDM does
	baseURL := "https://usher.ttvnw.net/api/channel/hls/" + channelLogin + ".m3u8"

	// Add query parameters like TDM
	params := fmt.Sprintf("?client_id=%s&token=%s&sig=%s&allow_source=true&allow_audio_only=true&allow_spectre=false&p=%d",
		g.clientInfo.ClientID,
		token.Value,
		token.Signature,
		generateRandomNumber())

	streamURL := baseURL + params
	logrus.Debugf("Generated stream URL for %s", channelLogin)

	return streamURL, nil
}

// SendWatchRequest sends a HEAD request to simulate watching (exactly like TDM)
func (g *GraphQLClient) SendWatchRequest(ctx context.Context, streamURL string) error {
	// Get the m3u8 playlist first
	req, err := http.NewRequestWithContext(ctx, "GET", streamURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create playlist request: %w", err)
	}

	// Set headers like TDM
	req.Header.Set("User-Agent", g.clientInfo.UserAgent)
	req.Header.Set("Client-ID", g.clientInfo.ClientID)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("playlist request failed with status: %d", resp.StatusCode)
	}

	// Read playlist content
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read playlist: %w", err)
	}

	// Parse m3u8 to find a stream playlist URL first
	playlistContent := string(body)
	logrus.Debugf("M3U8 master playlist received")

	// Extract a stream playlist URL (not chunk URL yet)
	streamPlaylistURL, err := g.extractStreamPlaylistURL(playlistContent)
	if err != nil {
		return fmt.Errorf("failed to extract stream playlist URL: %w", err)
	}

	// Now get the actual stream playlist with chunks
	chunkURL, err := g.getLastChunkFromPlaylist(ctx, streamPlaylistURL)
	if err != nil {
		return fmt.Errorf("failed to get chunk from stream playlist: %w", err)
	}

	// Send HEAD request to the chunk (this is what advances drops)
	headReq, err := http.NewRequestWithContext(ctx, "HEAD", chunkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create watch request: %w", err)
	}

	headReq.Header.Set("User-Agent", g.clientInfo.UserAgent)

	headResp, err := g.httpClient.Do(headReq)
	if err != nil {
		return fmt.Errorf("failed to send watch request: %w", err)
	}
	defer headResp.Body.Close()

	logrus.Debugf("Watch request sent, status: %d", headResp.StatusCode)
	return nil
}

// extractStreamPlaylistURL extracts a stream playlist URL from master playlist
func (g *GraphQLClient) extractStreamPlaylistURL(masterPlaylist string) (string, error) {
	lines := strings.Split(masterPlaylist, "\n")

	// Find any stream playlist URL (they end with .m3u8)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "http") && strings.Contains(line, ".m3u8") {
			return line, nil
		}
	}

	return "", fmt.Errorf("no stream playlist URL found in master playlist")
}

// getLastChunkFromPlaylist fetches a stream playlist and extracts the last chunk
func (g *GraphQLClient) getLastChunkFromPlaylist(ctx context.Context, playlistURL string) (string, error) {
	// Fetch the stream playlist
	req, err := http.NewRequestWithContext(ctx, "GET", playlistURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create stream playlist request: %w", err)
	}

	req.Header.Set("User-Agent", g.clientInfo.UserAgent)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get stream playlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("stream playlist request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read stream playlist: %w", err)
	}

	// Extract the last chunk from this playlist
	streamPlaylistContent := string(body)
	logrus.Debugf("Received stream playlist with %d lines", len(strings.Split(streamPlaylistContent, "\n")))
	return g.extractLastChunk(streamPlaylistContent, playlistURL)
}

// extractLastChunk extracts the last chunk URL from m3u8 playlist (like TDM)
func (g *GraphQLClient) extractLastChunk(playlist, baseURL string) (string, error) {
	lines := strings.Split(playlist, "\n")
	var lastChunkLine string

	// Find the last .ts file in the playlist (including query parameters)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Check if line contains .ts (might have query params after)
		if strings.Contains(line, ".ts") && strings.HasPrefix(line, "http") {
			lastChunkLine = line
		}
	}

	if lastChunkLine == "" {
		return "", fmt.Errorf("no chunk found in playlist (looked for .ts URLs)")
	}

	logrus.Debugf("Selected chunk URL: %s", lastChunkLine)
	return lastChunkLine, nil
}

// generateRandomNumber generates a random number like TDM does
func generateRandomNumber() int {
	// Simple random number for request uniqueness
	return int(time.Now().UnixNano() % 1000000)
}

// parseCampaignsResponse parses the campaigns GraphQL response
func (g *GraphQLClient) parseCampaignsResponse(data interface{}) ([]Campaign, error) {
	logrus.Debugf("Processing campaigns response")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	currentUser, ok := dataMap["currentUser"]
	if !ok || currentUser == nil {
		logrus.Warning("No currentUser in GraphQL response - authentication may be invalid")
		return []Campaign{}, nil
	}

	userMap, ok := currentUser.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid currentUser format")
	}

	// Look for dropCampaigns in the user data
	dropCampaigns, ok := userMap["dropCampaigns"]
	if !ok {
		logrus.Debug("No dropCampaigns found in currentUser")
		return []Campaign{}, nil
	}

	campaignsList, ok := dropCampaigns.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid dropCampaigns format - expected array")
	}

	var campaigns []Campaign
	for i, campaignInterface := range campaignsList {
		campaignMap, ok := campaignInterface.(map[string]interface{})
		if !ok {
			logrus.Errorf("Campaign %d: invalid campaign format", i)
			continue
		}

		campaign, err := g.parseCampaignNode(campaignMap)
		if err != nil {
			logrus.Errorf("Campaign %d: Failed to parse campaign: %v", i, err)
			continue
		}

		campaigns = append(campaigns, *campaign)
	}

	logrus.Debugf("Parsed %d campaigns from response", len(campaigns))
	return campaigns, nil
}

// GetCampaignDetails fetches detailed information about a specific campaign
func (g *GraphQLClient) GetCampaignDetails(ctx context.Context, campaignID string, userLogin string) (*Campaign, error) {
	resp, err := g.executeOperation(ctx, OpCampaignDetails, map[string]interface{}{
		"dropID":       campaignID,
		"channelLogin": userLogin,
	})
	if err != nil {
		return nil, err
	}

	// Parse detailed campaign from response
	campaign, err := g.parseCampaignDetailsResponse(resp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse campaign details: %w", err)
	}

	return campaign, nil
}

// parseCampaignDetailsResponse parses the campaign details response
func (g *GraphQLClient) parseCampaignDetailsResponse(data interface{}) (*Campaign, error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	user, ok := dataMap["user"]
	if !ok || user == nil {
		logrus.Warning("No user in CampaignDetails response")
		return nil, fmt.Errorf("no user in response")
	}

	userMap, ok := user.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid user format")
	}

	dropCampaign, ok := userMap["dropCampaign"]
	if !ok || dropCampaign == nil {
		return nil, fmt.Errorf("no dropCampaign in response")
	}

	campaignMap, ok := dropCampaign.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid dropCampaign format")
	}

	// Parse the detailed campaign
	return g.parseCampaignNode(campaignMap)
}

// parseInventoryResponse parses the inventory GraphQL response
func (g *GraphQLClient) parseInventoryResponse(data interface{}) (*InventoryGQL, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inventory data: %w", err)
	}

	var opResp OpInventoryResponse
	err = json.Unmarshal(dataBytes, &opResp)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling OpInventoryResponse: %w", err)
	}

	if opResp.CurrentUser == nil || opResp.CurrentUser.Inventory == nil {
		logrus.Warning("No currentUser or inventory in OpInventoryResponse")
		return &InventoryGQL{}, nil
	}

	return opResp.CurrentUser.Inventory, nil
}

// parseCampaignNode parses a single campaign node
func (g *GraphQLClient) parseCampaignNode(node map[string]interface{}) (*Campaign, error) {
	campaign := &Campaign{}

	// Basic campaign info
	campaign.ID = getString(node, "id")
	campaign.Name = getString(node, "name")
	campaign.Description = getString(node, "description")
	campaign.Status = getString(node, "status")
	campaign.ImageURL = getString(node, "imageURL")

	// Debug: Log campaign details for Don't Starve Together
	gameName := ""
	if game, ok := node["game"].(map[string]interface{}); ok {
		gameName = getString(game, "displayName")
	}

	// Game info
	if game, ok := node["game"].(map[string]interface{}); ok {
		campaign.Game = Game{
			ID:        getString(game, "id"),
			Name:      getString(game, "displayName"),
			BoxArtURL: getString(game, "boxArtURL"),
		}
	}

	// Time-based drops
	if timeBasedDrops, ok := node["timeBasedDrops"].([]interface{}); ok {
		logrus.Debugf("Campaign '%s' (%s): Found %d timeBasedDrops", campaign.Name, gameName, len(timeBasedDrops))
		for i, dropInterface := range timeBasedDrops {
			if dropMap, ok := dropInterface.(map[string]interface{}); ok {
				drop := TimeBased{
					ID:   getString(dropMap, "id"),
					Name: getString(dropMap, "name"),
				}

				if requiredMinutes, ok := dropMap["requiredMinutesWatched"].(float64); ok {
					drop.RequiredMinutesWatched = int(requiredMinutes)
				}

				logrus.Debugf("  Drop %d: '%s' requires %d minutes", i, drop.Name, drop.RequiredMinutesWatched)

				// Note: GetCampaignDetails response doesn't include user progress ("self" field)
				// User progress comes from DropCurrentSessionContext API call
				// Initialize with default values - progress will be updated separately
				drop.Self.IsClaimed = false
				drop.Self.CurrentMinutesWatched = 0
				drop.Self.DropInstanceID = ""

				logrus.Debugf("    Drop '%s' initialized with default progress (will be updated from DropCurrentSessionContext)", drop.Name)

				campaign.TimeBasedDrops = append(campaign.TimeBasedDrops, drop)
			} else {
				logrus.Errorf("Campaign '%s' (%s): Drop %d is not a valid map", campaign.Name, gameName, i)
			}
		}
	} else {
		// Check what type timeBasedDrops actually is
		if timeBasedDropsRaw, exists := node["timeBasedDrops"]; exists {
			logrus.Errorf("Campaign '%s' (%s): timeBasedDrops exists but wrong type: %T", campaign.Name, gameName, timeBasedDropsRaw)
		} else {
			logrus.Debugf("Campaign '%s' (%s): No timeBasedDrops field found", campaign.Name, gameName)
		}
	}

	// Campaign self info
	if self, ok := node["self"].(map[string]interface{}); ok {
		campaign.Self.IsAccountConnected = getBool(self, "isAccountConnected")
	}

	return campaign, nil
}

// parseStreamsResponse parses the streams GraphQL response
func (g *GraphQLClient) parseStreamsResponse(data interface{}) ([]Stream, error) {
	// Log the raw streams response to debug stream structure
	responseJSON, _ := json.MarshalIndent(data, "", "  ")
	logrus.Debugf("=== RAW STREAMS RESPONSE ===")
	logrus.Debugf("%s", string(responseJSON))
	logrus.Debugf("=== END STREAMS RESPONSE ===")

	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format")
	}

	game, ok := dataMap["game"]
	if !ok || game == nil {
		logrus.Debug("No game data in streams response")
		return []Stream{}, nil
	}

	gameMap, ok := game.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid game format")
	}

	streams, ok := gameMap["streams"]
	if !ok {
		logrus.Debug("No streams found in game data")
		return []Stream{}, nil
	}

	streamsMap, ok := streams.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid streams format")
	}

	edges, ok := streamsMap["edges"].([]interface{})
	if !ok {
		logrus.Debug("No stream edges found")
		return []Stream{}, nil
	}

	var streamList []Stream
	for _, edgeInterface := range edges {
		edgeMap, ok := edgeInterface.(map[string]interface{})
		if !ok {
			continue
		}

		node, ok := edgeMap["node"].(map[string]interface{})
		if !ok {
			continue
		}

		stream := Stream{
			ID:              getString(node, "id"),
			Title:           getString(node, "title"),
			PreviewImageURL: getString(node, "previewImageURL"),
		}

		if broadcaster, ok := node["broadcaster"].(map[string]interface{}); ok {
			stream.UserID = getString(broadcaster, "id")
			stream.UserLogin = getString(broadcaster, "login")
			stream.UserName = getString(broadcaster, "displayName")
		}

		if game, ok := node["game"].(map[string]interface{}); ok {
			stream.GameName = getString(game, "displayName")
		}

		if viewersCount, ok := node["viewersCount"].(float64); ok {
			stream.ViewerCount = int(viewersCount)
		}

		streamList = append(streamList, stream)
	}

	return streamList, nil
}

// Helper functions for safe type assertions
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBool(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
