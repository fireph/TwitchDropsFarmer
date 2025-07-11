package twitch

import "time"

// User represents a Twitch user
type User struct {
	ID              string `json:"id"`
	Login           string `json:"login"`
	DisplayName     string `json:"display_name"`
	Email           string `json:"email"`
	ProfileImageUrl string `json:"profile_image_url"`
}

// Game represents a Twitch game/category
type Game struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BoxArtURL string `json:"box_art_url"`
}

// Stream represents a live Twitch stream
type Stream struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	UserLogin       string    `json:"user_login"`
	UserName        string    `json:"user_name"`
	GameID          string    `json:"game_id"`
	GameName        string    `json:"game_name"`
	Type            string    `json:"type"`
	Title           string    `json:"title"`
	ViewerCount     int       `json:"viewer_count"`
	StartedAt       time.Time `json:"started_at"`
	Language        string    `json:"language"`
	PreviewImageURL string    `json:"preview_image_url"`
	TagIDs          []string  `json:"tag_ids"`
}

// Campaign represents a Twitch drop campaign
type Campaign struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Description    string       `json:"description"`
	Game           Game         `json:"game"`
	Status         string       `json:"status"`
	StartsAt       time.Time    `json:"starts_at"`
	EndsAt         time.Time    `json:"ends_at"`
	AccountLinkURL string       `json:"account_link_url"`
	Self           CampaignSelf `json:"self"`
	TimeBasedDrops []TimeBased  `json:"time_based_drops"`
	Allow          []string     `json:"allow"`
	Deny           []string     `json:"deny"`
	ImageURL       string       `json:"image_url"`
}

// CampaignSelf represents user's relationship to a campaign
type CampaignSelf struct {
	IsAccountConnected bool `json:"is_account_connected"`
}

// TimeBased represents a time-based drop
type TimeBased struct {
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	BenefitEdges           []BenefitEdge `json:"benefit_edges"`
	RequiredMinutesWatched int           `json:"required_minutes_watched"`
	Self                   TimeBasedSelf `json:"self"`
}

// TimeBasedSelf represents user's progress on a time-based drop
type TimeBasedSelf struct {
	CurrentMinutesWatched int    `json:"current_minutes_watched"`
	IsClaimed             bool   `json:"is_claimed"`
	DropInstanceID        string `json:"drop_instance_id"`
}

// BenefitEdge represents a drop benefit
type BenefitEdge struct {
	Benefit Benefit `json:"benefit"`
}

// Benefit represents a drop reward
type Benefit struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	ImageAssetURL string `json:"image_asset_url"`
	IsIOS         bool   `json:"is_ios"`
	IsAndroid     bool   `json:"is_android"`
	Game          Game   `json:"game"`
}

// PlaybackAccessToken represents stream access token
type PlaybackAccessToken struct {
	Value     string `json:"value"`
	Signature string `json:"signature"`
}

// WatchingSession represents an active stream watching session
type WatchingSession struct {
	ChannelLogin string
	StreamURL    string
	GQLClient    *GraphQLClient
}

// CurrentDropProgress represents current drop progress from TDM's CurrentDrop operation
type CurrentDropProgress struct {
	CurrentMinutesWatched int
	DropID                string
}

// GameSlugInfo represents the response from SlugRedirect/DirectoryGameRedirect
type GameSlugInfo struct {
	ID   string
	Slug string
}

// Inventory represents user's drop inventory
type Inventory struct {
	GameEventDrops []GameEventDrop `json:"game_event_drops"`
}

// GameEventDrop represents a claimed drop
type GameEventDrop struct {
	ID            string    `json:"id"`
	Benefit       Benefit   `json:"benefit"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
	Self          DropSelf  `json:"self"`
}

// DropSelf represents user's relationship to a drop
type DropSelf struct {
	IsClaimed bool `json:"is_claimed"`
}

// GraphQLResponse represents a GraphQL response
type GraphQLResponse struct {
	Data       interface{}            `json:"data"`
	Errors     []GraphQLError         `json:"errors,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation represents a GraphQL error location
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresIn    int      `json:"expires_in"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

// ViewerHeartbeatPayload represents WebSocket heartbeat payload
type ViewerHeartbeatPayload struct {
	Data struct {
		Topic string `json:"topic"`
		Type  string `json:"type"`
	} `json:"data"`
}

// MinerStatus represents the current status of the drop miner
type MinerStatus struct {
	IsRunning       bool      `json:"is_running"`
	CurrentStream   *Stream   `json:"current_stream"`
	CurrentCampaign *Campaign `json:"current_campaign"`
	CurrentProgress int       `json:"current_progress"`
	TotalCampaigns  int       `json:"total_campaigns"`
	ClaimedDrops    int       `json:"claimed_drops"`
	LastUpdate      time.Time `json:"last_update"`
	NextSwitch      time.Time `json:"next_switch"`
	ErrorMessage    string    `json:"error_message"`
}
