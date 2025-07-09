package storage

import "time"

// User represents a stored user
type User struct {
	ID           string    `json:"id"`
	Login        string    `json:"login"`
	DisplayName  string    `json:"display_name"`
	Email        string    `json:"email"`
	Avatar       string    `json:"avatar"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenExpiry  time.Time `json:"token_expiry"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Campaign represents a stored campaign
type Campaign struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	Description        string    `json:"description"`
	GameID             string    `json:"game_id"`
	GameName           string    `json:"game_name"`
	Status             string    `json:"status"`
	StartsAt           time.Time `json:"starts_at"`
	EndsAt             time.Time `json:"ends_at"`
	AccountLinkURL     string    `json:"account_link_url"`
	IsAccountConnected bool      `json:"is_account_connected"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Drop represents a stored drop
type Drop struct {
	ID              string    `json:"id"`
	CampaignID      string    `json:"campaign_id"`
	Name            string    `json:"name"`
	RequiredMinutes int       `json:"required_minutes"`
	CurrentMinutes  int       `json:"current_minutes"`
	IsClaimed       bool      `json:"is_claimed"`
	DropInstanceID  string    `json:"drop_instance_id"`
	BenefitID       string    `json:"benefit_id"`
	BenefitName     string    `json:"benefit_name"`
	BenefitImageURL string    `json:"benefit_image_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Stream represents a stored stream
type Stream struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	UserLogin     string    `json:"user_login"`
	UserName      string    `json:"user_name"`
	GameID        string    `json:"game_id"`
	GameName      string    `json:"game_name"`
	Title         string    `json:"title"`
	ViewerCount   int       `json:"viewer_count"`
	StartedAt     time.Time `json:"started_at"`
	Language      string    `json:"language"`
	ThumbnailURL  string    `json:"thumbnail_url"`
	IsWatching    bool      `json:"is_watching"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// MiningSession represents a stored mining session
type MiningSession struct {
	ID             int64     `json:"id"`
	UserID         string    `json:"user_id"`
	CampaignID     string    `json:"campaign_id"`
	StreamID       string    `json:"stream_id"`
	StartedAt      time.Time `json:"started_at"`
	EndedAt        time.Time `json:"ended_at"`
	MinutesWatched int       `json:"minutes_watched"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Settings represents a stored setting
type Settings struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
