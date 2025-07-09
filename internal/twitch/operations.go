package twitch

import (
	"encoding/json"
	"fmt"
)

// GQLOperation represents a GraphQL operation with persisted query support
type GQLOperation struct {
	OperationName string                 `json:"operationName"`
	Extensions    GQLExtensions          `json:"extensions"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
}

// GQLExtensions contains the persisted query information
type GQLExtensions struct {
	PersistedQuery GQLPersistedQuery `json:"persistedQuery"`
}

// GQLPersistedQuery contains the version and hash for the persisted query
type GQLPersistedQuery struct {
	Version    int    `json:"version"`
	SHA256Hash string `json:"sha256Hash"`
}

// NewGQLOperation creates a new GraphQL operation
func NewGQLOperation(name string, sha256Hash string, variables map[string]interface{}) *GQLOperation {
	op := &GQLOperation{
		OperationName: name,
		Extensions: GQLExtensions{
			PersistedQuery: GQLPersistedQuery{
				Version:    1,
				SHA256Hash: sha256Hash,
			},
		},
	}
	if variables != nil {
		op.Variables = variables
	}
	return op
}

// WithVariables creates a copy of the operation with merged variables
func (op *GQLOperation) WithVariables(variables map[string]interface{}) *GQLOperation {
	newOp := &GQLOperation{
		OperationName: op.OperationName,
		Extensions:    op.Extensions,
		Variables:     make(map[string]interface{}),
	}

	// Copy existing variables
	if op.Variables != nil {
		for k, v := range op.Variables {
			newOp.Variables[k] = v
		}
	}

	// Merge new variables
	for k, v := range variables {
		newOp.Variables[k] = v
	}

	return newOp
}

// ToJSON converts the operation to JSON bytes
func (op *GQLOperation) ToJSON() ([]byte, error) {
	return json.Marshal(op)
}

// GQL Operations - Exact copies from TDM constants.py
var GQLOperations = map[string]*GQLOperation{
	// returns stream information for a particular channel
	"GetStreamInfo": NewGQLOperation(
		"VideoPlayerStreamInfoOverlayChannel",
		"198492e0857f6aedead9665c81c5a06d67b25b58034649687124083ff288597d",
		map[string]interface{}{
			"channel": nil, // channel login - to be filled in
		},
	),

	// can be used to claim channel points
	"ClaimCommunityPoints": NewGQLOperation(
		"ClaimCommunityPoints",
		"46aaeebe02c99afdf4fc97c7c0cba964124bf6b0af229395f1f6d1feed05b3d0",
		map[string]interface{}{
			"input": map[string]interface{}{
				"claimID":   nil, // points claim_id - to be filled in
				"channelID": nil, // channel ID as a str - to be filled in
			},
		},
	),

	// can be used to claim a drop
	"ClaimDrop": NewGQLOperation(
		"DropsPage_ClaimDropRewards",
		"a455deea71bdc9015b78eb49f4acfbce8baa7ccbedd28e549bb025bd0f751930",
		map[string]interface{}{
			"input": map[string]interface{}{
				"dropInstanceID": nil, // drop claim_id - to be filled in
			},
		},
	),

	// returns current state of points (balance, claim available) for a particular channel
	"ChannelPointsContext": NewGQLOperation(
		"ChannelPointsContext",
		"374314de591e69925fce3ddc2bcf085796f56ebb8cad67a0daa3165c03adc345",
		map[string]interface{}{
			"channelLogin": nil, // channel login - to be filled in
		},
	),

	// returns all in-progress campaigns
	"Inventory": NewGQLOperation(
		"Inventory",
		"09acb7d3d7e605a92bdfdcc465f6aa481b71c234d8686a9ba38ea5ed51507592",
		map[string]interface{}{
			"fetchRewardCampaigns": false,
		},
	),

	// returns current state of drops (current drop progress)
	"CurrentDrop": NewGQLOperation(
		"DropCurrentSessionContext",
		"4d06b702d25d652afb9ef835d2a550031f1cf762b193523a92166f40ea3d142b",
		map[string]interface{}{
			"channelID":    nil, // watched channel ID as a str - to be filled in
			"channelLogin": "",  // always empty string
		},
	),

	// returns all available campaigns
	"Campaigns": NewGQLOperation(
		"ViewerDropsDashboard",
		"5a4da2ab3d5b47c9f9ce864e727b2cb346af1e3ea8b897fe8f704a97ff017619",
		map[string]interface{}{
			"fetchRewardCampaigns": false,
		},
	),

	// returns extended information about a particular campaign
	"CampaignDetails": NewGQLOperation(
		"DropCampaignDetails",
		"039277bf98f3130929262cc7c6efd9c141ca3749cb6dca442fc8ead9a53f77c1",
		map[string]interface{}{
			"channelLogin": nil, // user login - to be filled in
			"dropID":       nil, // campaign ID - to be filled in
		},
	),

	// returns drops available for a particular channel
	"AvailableDrops": NewGQLOperation(
		"DropsHighlightService_AvailableDrops",
		"9a62a09bce5b53e26e64a671e530bc599cb6aab1e5ba3cbd5d85966d3940716f",
		map[string]interface{}{
			"channelID": nil, // channel ID as a str - to be filled in
		},
	),

	// returns stream playback access token
	"PlaybackAccessToken": NewGQLOperation(
		"PlaybackAccessToken",
		"ed230aa1e33e07eebb8928504583da78a5173989fadfb1ac94be06a04f3cdbe9",
		map[string]interface{}{
			"isLive":     true,
			"isVod":      false,
			"login":      nil, // channel login - to be filled in
			"platform":   "web",
			"playerType": "site",
			"vodID":      "",
		},
	),

	// returns live channels for a particular game
	"GameDirectory": NewGQLOperation(
		"DirectoryPage_Game",
		"c7c9d5aad09155c4161d2382092dc44610367f3536aac39019ec2582ae5065f9",
		map[string]interface{}{
			"limit":       30,  // limit of channels returned
			"slug":        nil, // game slug - to be filled in
			"imageWidth":  50,
			"includeIsDJ": false,
			"options": map[string]interface{}{
				"broadcasterLanguages":   []interface{}{},
				"freeformTags":           nil,
				"includeRestricted":      []string{"SUB_ONLY_LIVE"},
				"recommendationsContext": map[string]interface{}{"platform": "web"},
				"sort":                   "RELEVANCE", // also accepted: "VIEWER_COUNT"
				"systemFilters":          []interface{}{},
				"tags":                   []interface{}{},
				"requestID":              "JIRA-VXP-2397",
			},
			"sortTypeIsRecency": false,
		},
	),

	// can be used to turn game name -> game slug
	"SlugRedirect": NewGQLOperation(
		"DirectoryGameRedirect",
		"1f0300090caceec51f33c5e20647aceff9017f740f223c3c532ba6fa59f6b6cc",
		map[string]interface{}{
			"name": nil, // game name - to be filled in
		},
	),

	// unused, triggers notifications "update-summary"
	"NotificationsView": NewGQLOperation(
		"OnsiteNotifications_View",
		"e8e06193f8df73d04a1260df318585d1bd7a7bb447afa058e52095513f2bfa4f",
		map[string]interface{}{
			"input": map[string]interface{}{},
		},
	),

	// unused
	"NotificationsList": NewGQLOperation(
		"OnsiteNotifications_ListNotifications",
		"11cdb54a2706c2c0b2969769907675680f02a6e77d8afe79a749180ad16bfea6",
		map[string]interface{}{
			"cursor":                  "",
			"displayType":             "VIEWER",
			"language":                "en",
			"limit":                   10,
			"shouldLoadLastBroadcast": false,
		},
	),

	"NotificationsDelete": NewGQLOperation(
		"OnsiteNotifications_DeleteNotification",
		"13d463c831f28ffe17dccf55b3148ed8b3edbbd0ebadd56352f1ff0160616816",
		map[string]interface{}{
			"input": map[string]interface{}{
				"id": "", // ID of the notification to delete
			},
		},
	),
}

// Helper function to get an operation and fill in variables
func GetOperation(name string, variables map[string]interface{}) (*GQLOperation, error) {
	baseOp, exists := GQLOperations[name]
	if !exists {
		return nil, fmt.Errorf("operation %s not found", name)
	}

	if variables == nil {
		return baseOp, nil
	}

	return baseOp.WithVariables(variables), nil
}
