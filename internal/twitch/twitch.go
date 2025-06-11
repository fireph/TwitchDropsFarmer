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
	// Search for games - from TwitchDropsMiner
	"SearchResultsPage_SearchResults": {
		OperationName: "SearchResultsPage_SearchResults",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "428d7d3d2a8e46ba88b3ed45c4c65203c5e2d1b26de3fe94c4e0e37b57c36853",
			},
		},
	},
	// Get live streams for a game directory - from TwitchDropsMiner
	"DirectoryPage_Game": {
		OperationName: "DirectoryPage_Game",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "df4bb6cc45055237bfaf3ead608bbafb79815c7100b6ee126719fac2a3924f8b",
			},
		},
	},
	// Get stream information for watching - from TwitchDropsMiner
	"VideoPlayerStreamInfoOverlayChannel": {
		OperationName: "VideoPlayerStreamInfoOverlayChannel",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "198492e0857f6aedead9665c81c5a06d67b25b58034649687124083ff288597d",
			},
		},
	},
	// Get drops inventory - from TwitchDropsMiner
	"Inventory": {
		OperationName: "Inventory",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "37fea1b41c6563007cc8a4327b90a52d42fda6b84950dfc17f42ad6bb8ea3a4b",
			},
		},
	},
	// Current drop progress - from TwitchDropsMiner
	"CurrentDrop": {
		OperationName: "DropCurrentSession",
		Extensions: &GQLExtensions{
			PersistedQuery: &PersistedQuery{
				Version:    1,
				SHA256Hash: "2e4b3630b91552eb05b76a94b6850eb25fe42263b7cf6d06bee6d156dd247c1c",
			},
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

// SearchGames searches for games by name using exact TwitchDropsMiner approach
func (c *Client) SearchGames(query string) ([]Game, error) {
	op := GQLOperations["SearchResultsPage_SearchResults"]
	op.Variables = map[string]interface{}{
		"query": query,
		"target": "GAME",
		"cursor": "",
	}

	fmt.Printf("Searching for games with query: %s\n", query)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return nil, fmt.Errorf("GraphQL search request failed: %v", err)
	}

	// Parse the search results exactly like TwitchDropsMiner would
	var games []Game
	if len(results) > 0 {
		data, ok := results[0]["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no data field")
		}
		
		searchResults, ok := data["searchResultsPage"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no searchResultsPage field")
		}
		
		edges, ok := searchResults["edges"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no edges field")
		}
		
		fmt.Printf("Found %d search results\n", len(edges))
		
		for _, edge := range edges {
			edgeMap, ok := edge.(map[string]interface{})
			if !ok {
				continue
			}
			
			node, ok := edgeMap["node"].(map[string]interface{})
			if !ok {
				continue
			}
			
			// Only include games, not other types
			if nodeType, ok := node["__typename"].(string); ok && nodeType != "Game" {
				continue
			}
			
			game := Game{
				ID:          getString(node, "id"),
				Name:        getString(node, "name"),
				DisplayName: getString(node, "displayName"),
				BoxArtURL:   getString(node, "boxArtURL"),
			}
			
			if game.ID != "" && game.DisplayName != "" {
				games = append(games, game)
			}
		}
	}
	
	fmt.Printf("Parsed %d games from search results\n", len(games))
	return games, nil
}

// GetStreamsForGame gets live streams for a specific game using exact TwitchDropsMiner approach
func (c *Client) GetStreamsForGame(gameID string, limit int) ([]Stream, error) {
	op := GQLOperations["DirectoryPage_Game"]
	op.Variables = map[string]interface{}{
		"name": gameID,
		"options": map[string]interface{}{
			"includeRestricted": []string{"SUB_ONLY_VIDEOS"},
			"sort":              "RELEVANCE",
			"recommendationsContext": map[string]interface{}{
				"platform": "web",
			},
			"requestID": generateRequestID(),
			"freeformTags": nil,
			"tags":         []string{},
		},
		"sortTypeIsRecency": false,
		"limit":            limit,
	}

	fmt.Printf("Getting streams for game ID: %s\n", gameID)

	results, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return nil, fmt.Errorf("GraphQL streams request failed: %v", err)
	}

	// Parse the stream results exactly like TwitchDropsMiner would
	var streams []Stream
	if len(results) > 0 {
		data, ok := results[0]["data"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no data field")
		}
		
		game, ok := data["game"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no game field")
		}
		
		streamsData, ok := game["streams"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no streams field")
		}
		
		edges, ok := streamsData["edges"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid response structure: no edges field")
		}
		
		fmt.Printf("Found %d stream results\n", len(edges))
		
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
				GameID:      gameID,
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

// WatchStream simulates watching a stream using exact TwitchDropsMiner approach
func (c *Client) WatchStream(channelLogin string) error {
	op := GQLOperations["VideoPlayerStreamInfoOverlayChannel"]
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

// GetDropCampaigns gets available drop campaigns using exact TwitchDropsMiner approach
func (c *Client) GetDropCampaigns() ([]Campaign, error) {
	op := GQLOperations["Inventory"]
	
	fmt.Println("Getting drop campaigns inventory")

	_, err := c.GraphQLRequest([]GQLOperation{op})
	if err != nil {
		return nil, fmt.Errorf("inventory request failed: %v", err)
	}

	// Parse the campaign results - this would need to be implemented
	// based on the actual response structure from Twitch
	var campaigns []Campaign
	// TODO: Implement proper response parsing based on Twitch's actual inventory GraphQL schema
	
	fmt.Printf("Found %d campaigns\n", len(campaigns))
	return campaigns, nil
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