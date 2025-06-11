package twitch

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Twitch API client
type Client struct {
	clientID     string
	clientSecret string
	httpClient   *http.Client
	accessToken  string
	deviceID     string
}

// GraphQL operation structures matching TwitchDropsMiner exactly
type GQLOperation struct {
	OperationName string                 `json:"operationName"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	Extensions    *GQLExtensions         `json:"extensions,omitempty"`
}

type GQLExtensions struct {
	PersistedQuery *PersistedQuery `json:"persistedQuery,omitempty"`
}

type PersistedQuery struct {
	Version    int    `json:"version"`
	SHA256Hash string `json:"sha256Hash"`
}

// Common structures
type Game struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	BoxArtURL   string `json:"boxArtURL"`
}

type Stream struct {
	ID           string `json:"id"`
	UserID       string `json:"userID"`
	UserLogin    string `json:"userLogin"`
	UserName     string `json:"userName"`
	GameID       string `json:"gameID"`
	GameName     string `json:"gameName"`
	Title        string `json:"title"`
	ViewerCount  int    `json:"viewerCount"`
	StartedAt    string `json:"startedAt"`
	Language     string `json:"language"`
	ThumbnailURL string `json:"thumbnailURL"`
	TagIDs       []string `json:"tagIDs"`
	IsMature     bool   `json:"isMature"`
}

type Drop struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ImageURL      string    `json:"imageURL"`
	StartAt       time.Time `json:"startAt"`
	EndAt         time.Time `json:"endAt"`
	RequiredMinutes int     `json:"requiredMinutes"`
	CurrentMinutes  int     `json:"currentMinutes"`
	GameID        string    `json:"gameID"`
	GameName      string    `json:"gameName"`
	IsClaimed     bool      `json:"isClaimed"`
	IsCompleted   bool      `json:"isCompleted"`
}

type Campaign struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	GameID      string  `json:"gameID"`
	GameName    string  `json:"gameName"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	Drops       []Drop  `json:"drops"`
	Status      string  `json:"status"`
}

// OAuth2 response structure for SmartTV flow
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

// Exact GQL Operations from TwitchDropsMiner constants.py
var GQLOperations = map[string]GQLOperation{
	"GetStreamInfo": {
		OperationName: "VideoPlayerStreamInfoOverlayChannel",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "198492e0857f6aedead9665c81c5a06d67b25b58034649687124083ff288597d",
			},
		},
		Variables: map[string]interface{}{
			"channel": "", // channel login
		},
	},
	"ClaimCommunityPoints": {
		OperationName: "ClaimCommunityPoints",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "46aaeebe02c99afdf4fc97c7c0cba964124bf6b0af229395f1f6d1feed05b3d0",
			},
		},
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"claimID": "", // points claim_id
				"channelID": "", // channel ID as a str
			},
		},
	},
	"ClaimDrop": {
		OperationName: "DropsPage_ClaimDropRewards",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "a455deea71bdc9015b78eb49f4acfbce8baa7ccbedd28e549bb025bd0f751930",
			},
		},
		Variables: map[string]interface{}{
			"input": map[string]interface{}{
				"dropInstanceID": "", // drop claim_id
			},
		},
	},
	"ChannelPointsContext": {
		OperationName: "ChannelPointsContext",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "374314de591e69925fce3ddc2bcf085796f56ebb8cad67a0daa3165c03adc345",
			},
		},
		Variables: map[string]interface{}{
			"channelLogin": "", // channel login
		},
	},
	"Inventory": {
		OperationName: "Inventory",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "09acb7d3d7e605a92bdfdcc465f6aa481b71c234d8686a9ba38ea5ed51507592",
			},
		},
		Variables: map[string]interface{}{
			"fetchRewardCampaigns": false,
		},
	},
	"CurrentDrop": {
		OperationName: "DropCurrentSessionContext",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "4d06b702d25d652afb9ef835d2a550031f1cf762b193523a92166f40ea3d142b",
			},
		},
		Variables: map[string]interface{}{
			"channelID": "", // watched channel ID as a str
			"channelLogin": "", // always empty string
		},
	},
	"Campaigns": {
		OperationName: "ViewerDropsDashboard",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "5a4da2ab3d5b47c9f9ce864e727b2cb346af1e3ea8b897fe8f704a97ff017619",
			},
		},
		Variables: map[string]interface{}{
			"fetchRewardCampaigns": false,
		},
	},
	"CampaignDetails": {
		OperationName: "DropCampaignDetails",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "039277bf98f3130929262cc7c6efd9c141ca3749cb6dca442fc8ead9a53f77c1",
			},
		},
		Variables: map[string]interface{}{
			"channelLogin": "", // user login
			"dropID": "", // campaign ID
		},
	},
	"AvailableDrops": {
		OperationName: "DropsHighlightService_AvailableDrops",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "9a62a09bce5b53e26e64a671e530bc599cb6aab1e5ba3cbd5d85966d3940716f",
			},
		},
		Variables: map[string]interface{}{
			"channelID": "", // channel ID as a str
		},
	},
	"PlaybackAccessToken": {
		OperationName: "PlaybackAccessToken",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "ed230aa1e33e07eebb8928504583da78a5173989fadfb1ac94be06a04f3cdbe9",
			},
		},
		Variables: map[string]interface{}{
			"isLive": true,
			"isVod": false,
			"login": "", // channel login
			"platform": "web",
			"playerType": "site",
			"vodID": "",
		},
	},
	"GameDirectory": {
		OperationName: "DirectoryPage_Game",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "df4bb6cc45055237bfaf3ead608bbafb79815c7100b6ee126719fac2a3924f8b",
			},
		},
		Variables: map[string]interface{}{
			"limit": 30, // limit of channels returned
			"slug": "", // game slug
			"imageWidth": 50,
			"includeIsDJ": false,
			"options": map[string]interface{}{
				"broadcasterLanguages": []interface{}{},
				"freeformTags": nil,
				"includeRestricted": []interface{}{"SUB_ONLY_LIVE"},
				"recommendationsContext": map[string]interface{}{"platform": "web"},
				"sort": "RELEVANCE", // also accepted: "VIEWER_COUNT"
				"systemFilters": []interface{}{},
				"tags": []interface{}{},
				"requestID": "JIRA-VXP-2397",
			},
			"sortTypeIsRecency": false,
		},
	},
	"SlugRedirect": {
		OperationName: "DirectoryGameRedirect",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "1f0300090caceec51f33c5e20647aceff9017f740f223c3c532ba6fa59f6b6cc",
			},
		},
		Variables: map[string]interface{}{
			"name": "", // game name
		},
	},
}

// New creates a new Twitch client
func New(clientID, clientSecret string) *Client {
	deviceID := generateDeviceID()
	
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		deviceID:     deviceID,
	}
}

// generateDeviceID creates a device ID similar to TwitchDropsMiner
func generateDeviceID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// StartDeviceFlow initiates the SmartTV OAuth flow
func (c *Client) StartDeviceFlow() (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {c.clientID},
		"scopes":    {""},
	}

	req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/device", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("X-Device-Id", c.deviceID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PollForToken polls for the OAuth token
func (c *Client) PollForToken(deviceCode string, interval int) (*TokenResponse, error) {
	data := url.Values{
		"client_id":    {c.clientID},
		"device_code":  {deviceCode},
		"grant_type":   {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	maxAttempts := 60 // 5 minutes with 5-second intervals
	attempts := 0

	for attempts < maxAttempts {
		attempts++
		time.Sleep(time.Duration(interval) * time.Second)

		req, err := http.NewRequest("POST", "https://id.twitch.tv/oauth2/token", strings.NewReader(data.Encode()))
		if err != nil {
			fmt.Printf("Error creating request (attempt %d): %v\n", attempts, err)
			continue
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Client-ID", c.clientID)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			fmt.Printf("Error making request (attempt %d): %v\n", attempts, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		
		if err != nil {
			fmt.Printf("Error reading response (attempt %d): %v\n", attempts, err)
			continue
		}

		fmt.Printf("Auth polling attempt %d, status: %d, body: %s\n", attempts, resp.StatusCode, string(body))

		if resp.StatusCode == 200 {
			var result TokenResponse
			if err := json.Unmarshal(body, &result); err != nil {
				fmt.Printf("Error parsing success response: %v\n", err)
				return nil, err
			}
			
			fmt.Printf("Parsed token response: %+v\n", result)
			
			// If ExpiresIn is 0, set a default of 4 hours (typical for Twitch tokens)
			if result.ExpiresIn == 0 {
				fmt.Println("ExpiresIn is 0, setting default of 4 hours")
				result.ExpiresIn = 14400 // 4 hours
			}
			
			c.accessToken = result.AccessToken
			fmt.Println("Successfully obtained access token!")
			return &result, nil
		}

		if resp.StatusCode == 400 {
			// Parse error response to see what's happening
			var errorResp map[string]interface{}
			if err := json.Unmarshal(body, &errorResp); err == nil {
				if errorMsg, ok := errorResp["error"].(string); ok {
					fmt.Printf("Auth error (attempt %d): %s\n", attempts, errorMsg)
					if errorMsg == "authorization_pending" {
						// This is expected, continue polling
						continue
					} else if errorMsg == "slow_down" {
						// Increase interval
						time.Sleep(time.Duration(interval) * time.Second)
						continue
					} else if errorMsg == "expired_token" {
						return nil, errors.New("device code expired")
					} else if errorMsg == "access_denied" {
						return nil, errors.New("user denied authorization")
					}
				}
			}
			continue
		}

		fmt.Printf("Unexpected status code (attempt %d): %d\n", attempts, resp.StatusCode)
	}

	return nil, errors.New("authentication timed out after maximum attempts")
}

// SetAccessToken sets the access token for API calls
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

// GraphQLRequest makes a GraphQL request to Twitch using exact TwitchDropsMiner format
func (c *Client) GraphQLRequest(operations []GQLOperation) ([]map[string]interface{}, error) {
	var payload interface{}
	if len(operations) == 1 {
		payload = operations[0]
	} else {
		payload = operations
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	fmt.Printf("GraphQL Request: %s\n", string(jsonData))

	req, err := http.NewRequest("POST", "https://gql.twitch.tv/gql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// Use exact headers from TwitchDropsMiner
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Client-ID", c.clientID)
	req.Header.Set("X-Device-Id", c.deviceID)
	
	if c.accessToken != "" {
		req.Header.Set("Authorization", "OAuth "+c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("GraphQL Response Status: %d\n", resp.StatusCode)
	fmt.Printf("GraphQL Response Body: %s\n", string(body))

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GraphQL request failed: %s", string(body))
	}

	var result []map[string]interface{}
	if len(operations) == 1 {
		var single map[string]interface{}
		if err := json.Unmarshal(body, &single); err != nil {
			return nil, err
		}
		result = []map[string]interface{}{single}
	} else {
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// ResolveGameSlug resolves a game name to its slug using SlugRedirect
func (c *Client) ResolveGameSlug(gameName string) (string, error) {
	op := GQLOperations["SlugRedirect"]
	// Set the required name variable
	op.Variables = map[string]interface{}{
		"name": gameName,
	}

	fmt.Printf("Resolving game slug for: %s\n", gameName)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return "", fmt.Errorf("slug redirect request failed: %v", err)
	}
	
	// Parse the response to get the slug
	if len(results) > 0 {
		fmt.Printf("Slug redirect response: %+v\n", results[0])
		
		data, ok := results[0]["data"].(map[string]interface{})
		if ok {
			if game, hasGame := data["game"].(map[string]interface{}); hasGame {
				if slug, hasSlug := game["slug"].(string); hasSlug {
					fmt.Printf("Resolved game slug: %s -> %s\n", gameName, slug)
					return slug, nil
				}
			}
		}
	}
	
	return "", fmt.Errorf("could not resolve slug for game: %s", gameName)
}

// AddGame adds a game to the watch list with proper slug resolution
func (c *Client) AddGame(gameName string) (*Game, error) {
	// For unknown games, try to resolve the slug first
	slug, err := c.ResolveGameSlug(gameName)
	if err != nil {
		fmt.Printf("Could not resolve slug for %s, using name as-is: %v\n", gameName, err)
		slug = gameName
	}
	
	// Create game with resolved slug
	game := &Game{
		ID:          slug,  // Use the resolved slug as ID
		Name:        gameName,
		DisplayName: gameName,
		BoxArtURL:   "",
	}
	
	fmt.Printf("Added game: %s (slug: %s)\n", gameName, slug)
	return game, nil
}

// GetStreamsForGame gets live streams for a specific game using correct operation and variables
func (c *Client) GetStreamsForGame(gameNameOrID string, limit int) ([]Stream, error) {
	// Use the correct operation name from your GQLOperations
	op := GQLOperations["GameDirectory"]
	
	// Set the required variables according to your GQLOperations definition
	op.Variables = map[string]interface{}{
		"limit":     limit,
		"slug":      gameNameOrID, // Use slug instead of name
		"imageWidth": 50,
		"includeIsDJ": false,
		"options": map[string]interface{}{
			"broadcasterLanguages": []interface{}{},
			"freeformTags":         nil,
			"includeRestricted":    []interface{}{"SUB_ONLY_LIVE"},
			"recommendationsContext": map[string]interface{}{
				"platform": "web",
			},
			"sort":          "RELEVANCE",
			"systemFilters": []interface{}{},
			"tags":          []interface{}{},
			"requestID":     "JIRA-VXP-2397",
		},
		"sortTypeIsRecency": false,
	}

	fmt.Printf("Getting streams for game: %s\n", gameNameOrID)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return nil, fmt.Errorf("GraphQL streams request failed: %v", err)
	}

	// Parse the stream results
	var streams []Stream
	if len(results) > 0 {
		data, ok := results[0]["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no data field")
		}
		
		game, ok := data["game"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("game not found or invalid game name/ID: %s", gameNameOrID)
		}
		
		streamsData, ok := game["streams"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("no streams data found for game: %s", gameNameOrID)
		}
		
		edges, ok := streamsData["edges"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid streams structure: no edges field")
		}
		
		fmt.Printf("Found %d stream results for %s\n", len(edges), gameNameOrID)
		
		// Extract the actual game ID from the response for future use
		if gameID, ok := game["id"].(string); ok {
			fmt.Printf("Resolved game ID: %s -> %s\n", gameNameOrID, gameID)
		}
		
		for _, edge := range edges {
			edgeMap, ok := edge.(map[string]interface{})
			if !ok {
				continue
			}
			
			node, ok := edgeMap["node"].(map[string]interface{})
			if !ok {
				continue
			}
			
			broadcaster, ok := node["broadcaster"].(map[string]interface{})
			if !ok {
				continue
			}
			
			stream := Stream{
				ID:          getString(node, "id"),
				UserID:      getString(broadcaster, "id"),
				UserLogin:   getString(broadcaster, "login"),
				UserName:    getString(broadcaster, "displayName"),
				GameID:      getString(game, "id"),
				GameName:    getString(game, "displayName"),
				Title:       getString(node, "title"),
				ViewerCount: getInt(node, "viewersCount"),
				Language:    getString(broadcaster, "broadcastSettings.language"),
			}
			
			if stream.ID != "" && stream.UserLogin != "" {
				streams = append(streams, stream)
			}
		}
	}
	
	fmt.Printf("Parsed %d streams from results\n", len(streams))
	return streams, nil
}

// WatchStream simulates watching a stream using the correct operation
func (c *Client) WatchStream(channelLogin string) error {
	op := GQLOperations["GetStreamInfo"]
	// Set the required channel variable
	op.Variables = map[string]interface{}{
		"channel": channelLogin,
	}

	fmt.Printf("Watching stream: %s\n", channelLogin)

	_, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("watch request failed: %v", err)
	}
	
	fmt.Printf("Successfully watched stream: %s\n", channelLogin)
	return nil
}

// GetDropCampaigns gets available drop campaigns using the correct operations
func (c *Client) GetDropCampaigns() ([]Campaign, error) {
	fmt.Println("Getting drop campaigns")

	// Try the Campaigns operation first (ViewerDropsDashboard)
	op := GQLOperations["Campaigns"]
	// Set the required variable
	op.Variables = map[string]interface{}{
		"fetchRewardCampaigns": false,
	}
	
	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		if strings.Contains(err.Error(), "PersistedQueryNotFound") {
			fmt.Println("Campaigns operation failed, trying Inventory...")
			
			// Try Inventory operation as fallback
			inventoryOp := GQLOperations["Inventory"]
			inventoryOp.Variables = map[string]interface{}{
				"fetchRewardCampaigns": false,
			}
			
			results, err = c.GraphQLRequest([]GQLOperation{inventoryOp})
			if err != nil {
				return nil, fmt.Errorf("both campaign operations failed: %v", err)
			}
		} else {
			return nil, fmt.Errorf("campaigns request failed: %v", err)
		}
	}

	// Parse the campaign results
	var campaigns []Campaign
	
	// Log the response structure for debugging
	if len(results) > 0 {
		fmt.Printf("Campaigns response structure: %+v\n", results[0])
		
		// Try to parse the actual structure
		data, ok := results[0]["data"].(map[string]interface{})
		if ok {
			// Check for viewer data (ViewerDropsDashboard structure)
			if viewer, hasViewer := data["currentUser"].(map[string]interface{}); hasViewer {
				if dropCampaigns, hasCampaigns := viewer["dropCampaigns"].([]interface{}); hasCampaigns {
					fmt.Printf("Found %d campaigns in viewer data\n", len(dropCampaigns))
					// TODO: Parse campaigns from dropCampaigns array
				}
			}
			
			// Check for inventory data (Inventory structure)  
			if inventory, hasInventory := data["currentUser"].(map[string]interface{}); hasInventory {
				if invItems, hasItems := inventory["inventory"].(map[string]interface{}); hasItems {
					fmt.Printf("Found inventory data: %+v\n", invItems)
					// TODO: Parse campaigns from inventory
				}
			}
		}
	}
	
	fmt.Printf("Found %d campaigns\n", len(campaigns))
	return campaigns, nil
}

// GetCurrentDrop gets the current drop progress for a channel
func (c *Client) GetCurrentDrop(channelID, channelLogin string) error {
	op := GQLOperations["CurrentDrop"]
	// Set the required variables
	op.Variables = map[string]interface{}{
		"channelID":    channelID,
		"channelLogin": channelLogin, // or empty string as per your definition
	}

	fmt.Printf("Getting current drop for channel: %s\n", channelLogin)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("current drop request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Current drop response: %+v\n", results[0])
	}
	
	return nil
}

// ClaimDrop claims a drop using the correct operation
func (c *Client) ClaimDrop(dropInstanceID string) error {
	op := GQLOperations["ClaimDrop"]
	// Set the required input variable
	op.Variables = map[string]interface{}{
		"input": map[string]interface{}{
			"dropInstanceID": dropInstanceID,
		},
	}

	fmt.Printf("Claiming drop: %s\n", dropInstanceID)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("claim drop request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Claim drop response: %+v\n", results[0])
	}
	
	fmt.Printf("Successfully claimed drop: %s\n", dropInstanceID)
	return nil
}

// GetAvailableDrops gets available drops for a channel
func (c *Client) GetAvailableDrops(channelID string) error {
	op := GQLOperations["AvailableDrops"]
	// Set the required channelID variable
	op.Variables = map[string]interface{}{
		"channelID": channelID,
	}

	fmt.Printf("Getting available drops for channel ID: %s\n", channelID)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("available drops request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Available drops response: %+v\n", results[0])
	}
	
	return nil
}

// ClaimCommunityPoints claims community points for a channel
func (c *Client) ClaimCommunityPoints(claimID, channelID string) error {
	op := GQLOperations["ClaimCommunityPoints"]
	// Set the required input variables
	op.Variables = map[string]interface{}{
		"input": map[string]interface{}{
			"claimID":   claimID,
			"channelID": channelID,
		},
	}

	fmt.Printf("Claiming community points: %s for channel %s\n", claimID, channelID)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("claim community points request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Claim community points response: %+v\n", results[0])
	}
	
	fmt.Printf("Successfully claimed community points: %s\n", claimID)
	return nil
}

// GetChannelPointsContext gets channel points context for a channel
func (c *Client) GetChannelPointsContext(channelLogin string) error {
	op := GQLOperations["ChannelPointsContext"]
	// Set the required channelLogin variable
	op.Variables = map[string]interface{}{
		"channelLogin": channelLogin,
	}

	fmt.Printf("Getting channel points context for: %s\n", channelLogin)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("channel points context request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Channel points context response: %+v\n", results[0])
	}
	
	return nil
}

// GetCampaignDetails gets detailed information about a specific drop campaign
func (c *Client) GetCampaignDetails(channelLogin, dropID string) error {
	op := GQLOperations["CampaignDetails"]
	// Set the required variables
	op.Variables = map[string]interface{}{
		"channelLogin": channelLogin,
		"dropID":       dropID,
	}

	fmt.Printf("Getting campaign details for drop %s on channel %s\n", dropID, channelLogin)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("campaign details request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Campaign details response: %+v\n", results[0])
	}
	
	return nil
}

// GetPlaybackAccessToken gets a playback access token for a channel
func (c *Client) GetPlaybackAccessToken(channelLogin string) error {
	op := GQLOperations["PlaybackAccessToken"]
	// Set the required variables
	op.Variables = map[string]interface{}{
		"isLive":     true,
		"isVod":      false,
		"login":      channelLogin,
		"platform":   "web",
		"playerType": "site",
		"vodID":      "",
	}

	fmt.Printf("Getting playback access token for: %s\n", channelLogin)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return fmt.Errorf("playback access token request failed: %v", err)
	}
	
	// Log the response for debugging
	if len(results) > 0 {
		fmt.Printf("Playback access token response: %+v\n", results[0])
	}
	
	return nil
}

// Helper functions for parsing JSON responses
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

func getInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}

func generateRequestID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}