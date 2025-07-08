package storage

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	database := &Database{db: db}
	if err := database.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return database, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			login TEXT NOT NULL,
			display_name TEXT NOT NULL,
			email TEXT,
			avatar TEXT,
			access_token TEXT,
			refresh_token TEXT,
			token_expiry DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS campaigns (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			game_id TEXT NOT NULL,
			game_name TEXT NOT NULL,
			status TEXT NOT NULL,
			starts_at DATETIME,
			ends_at DATETIME,
			account_link_url TEXT,
			is_account_connected BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS drops (
			id TEXT PRIMARY KEY,
			campaign_id TEXT NOT NULL,
			name TEXT NOT NULL,
			required_minutes INTEGER NOT NULL,
			current_minutes INTEGER DEFAULT 0,
			is_claimed BOOLEAN DEFAULT FALSE,
			drop_instance_id TEXT,
			benefit_id TEXT,
			benefit_name TEXT,
			benefit_image_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (campaign_id) REFERENCES campaigns(id)
		)`,
		`CREATE TABLE IF NOT EXISTS streams (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			user_login TEXT NOT NULL,
			user_name TEXT NOT NULL,
			game_id TEXT NOT NULL,
			game_name TEXT NOT NULL,
			title TEXT NOT NULL,
			viewer_count INTEGER DEFAULT 0,
			started_at DATETIME,
			language TEXT,
			thumbnail_url TEXT,
			is_watching BOOLEAN DEFAULT FALSE,
			last_heartbeat DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS mining_sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			campaign_id TEXT NOT NULL,
			stream_id TEXT NOT NULL,
			started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			ended_at DATETIME,
			minutes_watched INTEGER DEFAULT 0,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (campaign_id) REFERENCES campaigns(id),
			FOREIGN KEY (stream_id) REFERENCES streams(id)
		)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_campaigns_game_id ON campaigns(game_id)`,
		`CREATE INDEX IF NOT EXISTS idx_drops_campaign_id ON drops(campaign_id)`,
		`CREATE INDEX IF NOT EXISTS idx_streams_game_id ON streams(game_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mining_sessions_user_id ON mining_sessions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_mining_sessions_campaign_id ON mining_sessions(campaign_id)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	return nil
}

// User methods
func (d *Database) SaveUser(user *User) error {
	query := `INSERT OR REPLACE INTO users (id, login, display_name, email, avatar, access_token, refresh_token, token_expiry, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := d.db.Exec(query, user.ID, user.Login, user.DisplayName, user.Email, user.Avatar, 
		user.AccessToken, user.RefreshToken, user.TokenExpiry, time.Now())
	return err
}

func (d *Database) GetUser(userID string) (*User, error) {
	query := `SELECT id, login, display_name, email, avatar, access_token, refresh_token, token_expiry 
			  FROM users WHERE id = ?`
	
	var user User
	err := d.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Login, &user.DisplayName, &user.Email, &user.Avatar,
		&user.AccessToken, &user.RefreshToken, &user.TokenExpiry,
	)
	if err != nil {
		return nil, err
	}
	
	return &user, nil
}

// Campaign methods
func (d *Database) SaveCampaign(campaign *Campaign) error {
	query := `INSERT OR REPLACE INTO campaigns (id, name, description, game_id, game_name, status, starts_at, ends_at, account_link_url, is_account_connected, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := d.db.Exec(query, campaign.ID, campaign.Name, campaign.Description, campaign.GameID, campaign.GameName,
		campaign.Status, campaign.StartsAt, campaign.EndsAt, campaign.AccountLinkURL, campaign.IsAccountConnected, time.Now())
	return err
}

func (d *Database) GetActiveCampaigns() ([]Campaign, error) {
	query := `SELECT id, name, description, game_id, game_name, status, starts_at, ends_at, account_link_url, is_account_connected 
			  FROM campaigns WHERE status = 'ACTIVE' AND starts_at <= ? AND ends_at >= ?`
	
	now := time.Now()
	rows, err := d.db.Query(query, now, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []Campaign
	for rows.Next() {
		var campaign Campaign
		err := rows.Scan(&campaign.ID, &campaign.Name, &campaign.Description, &campaign.GameID, &campaign.GameName,
			&campaign.Status, &campaign.StartsAt, &campaign.EndsAt, &campaign.AccountLinkURL, &campaign.IsAccountConnected)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

// Drop methods
func (d *Database) SaveDrop(drop *Drop) error {
	query := `INSERT OR REPLACE INTO drops (id, campaign_id, name, required_minutes, current_minutes, is_claimed, drop_instance_id, benefit_id, benefit_name, benefit_image_url, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := d.db.Exec(query, drop.ID, drop.CampaignID, drop.Name, drop.RequiredMinutes, drop.CurrentMinutes,
		drop.IsClaimed, drop.DropInstanceID, drop.BenefitID, drop.BenefitName, drop.BenefitImageURL, time.Now())
	return err
}

func (d *Database) GetDropsForCampaign(campaignID string) ([]Drop, error) {
	query := `SELECT id, campaign_id, name, required_minutes, current_minutes, is_claimed, drop_instance_id, benefit_id, benefit_name, benefit_image_url 
			  FROM drops WHERE campaign_id = ?`
	
	rows, err := d.db.Query(query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drops []Drop
	for rows.Next() {
		var drop Drop
		err := rows.Scan(&drop.ID, &drop.CampaignID, &drop.Name, &drop.RequiredMinutes, &drop.CurrentMinutes,
			&drop.IsClaimed, &drop.DropInstanceID, &drop.BenefitID, &drop.BenefitName, &drop.BenefitImageURL)
		if err != nil {
			return nil, err
		}
		drops = append(drops, drop)
	}

	return drops, nil
}

func (d *Database) UpdateDropProgress(dropID string, currentMinutes int) error {
	query := `UPDATE drops SET current_minutes = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, currentMinutes, time.Now(), dropID)
	return err
}

func (d *Database) MarkDropClaimed(dropID string) error {
	query := `UPDATE drops SET is_claimed = TRUE, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), dropID)
	return err
}

// Stream methods
func (d *Database) SaveStream(stream *Stream) error {
	query := `INSERT OR REPLACE INTO streams (id, user_id, user_login, user_name, game_id, game_name, title, viewer_count, started_at, language, thumbnail_url, updated_at) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	_, err := d.db.Exec(query, stream.ID, stream.UserID, stream.UserLogin, stream.UserName, stream.GameID, stream.GameName,
		stream.Title, stream.ViewerCount, stream.StartedAt, stream.Language, stream.ThumbnailURL, time.Now())
	return err
}

func (d *Database) GetStreamsForGame(gameID string) ([]Stream, error) {
	query := `SELECT id, user_id, user_login, user_name, game_id, game_name, title, viewer_count, started_at, language, thumbnail_url 
			  FROM streams WHERE game_id = ? ORDER BY viewer_count DESC`
	
	rows, err := d.db.Query(query, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var streams []Stream
	for rows.Next() {
		var stream Stream
		err := rows.Scan(&stream.ID, &stream.UserID, &stream.UserLogin, &stream.UserName, &stream.GameID, &stream.GameName,
			&stream.Title, &stream.ViewerCount, &stream.StartedAt, &stream.Language, &stream.ThumbnailURL)
		if err != nil {
			return nil, err
		}
		streams = append(streams, stream)
	}

	return streams, nil
}

func (d *Database) SetCurrentWatchingStream(streamID string) error {
	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Clear all current watching flags
	if _, err := tx.Exec(`UPDATE streams SET is_watching = FALSE`); err != nil {
		return err
	}

	// Set current stream as watching
	if _, err := tx.Exec(`UPDATE streams SET is_watching = TRUE WHERE id = ?`, streamID); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *Database) GetCurrentWatchingStream() (*Stream, error) {
	query := `SELECT id, user_id, user_login, user_name, game_id, game_name, title, viewer_count, started_at, language, thumbnail_url 
			  FROM streams WHERE is_watching = TRUE LIMIT 1`
	
	var stream Stream
	err := d.db.QueryRow(query).Scan(&stream.ID, &stream.UserID, &stream.UserLogin, &stream.UserName, &stream.GameID, &stream.GameName,
		&stream.Title, &stream.ViewerCount, &stream.StartedAt, &stream.Language, &stream.ThumbnailURL)
	if err != nil {
		return nil, err
	}
	
	return &stream, nil
}

// Mining session methods
func (d *Database) StartMiningSession(userID, campaignID, streamID string) (int64, error) {
	query := `INSERT INTO mining_sessions (user_id, campaign_id, stream_id, started_at, status) 
			  VALUES (?, ?, ?, ?, 'active')`
	
	result, err := d.db.Exec(query, userID, campaignID, streamID, time.Now())
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

func (d *Database) EndMiningSession(sessionID int64, minutesWatched int) error {
	query := `UPDATE mining_sessions SET ended_at = ?, minutes_watched = ?, status = 'completed', updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, time.Now(), minutesWatched, time.Now(), sessionID)
	return err
}

func (d *Database) GetActiveMiningSession(userID string) (*MiningSession, error) {
	query := `SELECT id, user_id, campaign_id, stream_id, started_at, minutes_watched, status 
			  FROM mining_sessions WHERE user_id = ? AND status = 'active' LIMIT 1`
	
	var session MiningSession
	err := d.db.QueryRow(query, userID).Scan(&session.ID, &session.UserID, &session.CampaignID, &session.StreamID,
		&session.StartedAt, &session.MinutesWatched, &session.Status)
	if err != nil {
		return nil, err
	}
	
	return &session, nil
}

// Settings methods
func (d *Database) SaveSetting(key, value string) error {
	query := `INSERT OR REPLACE INTO settings (key, value, updated_at) VALUES (?, ?, ?)`
	_, err := d.db.Exec(query, key, value, time.Now())
	return err
}

func (d *Database) GetSetting(key string) (string, error) {
	query := `SELECT value FROM settings WHERE key = ?`
	var value string
	err := d.db.QueryRow(query, key).Scan(&value)
	return value, err
}