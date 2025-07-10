// User represents a Twitch user
export interface User {
  id: string;
  login: string;
  display_name: string;
  email: string;
  profile_image_url: string;
}

// Game represents a Twitch game/category
export interface Game {
  id: string;
  name: string;
  box_art_url: string;
}

// Stream represents a live Twitch stream
export interface Stream {
  id: string;
  user_id: string;
  user_login: string;
  user_name: string;
  game_id: string;
  game_name: string;
  type: string;
  title: string;
  viewer_count: number;
  started_at: string; // ISO date string
  language: string;
  preview_image_url: string;
  tag_ids: string[];
}

// Campaign represents a Twitch drop campaign
export interface Campaign {
  id: string;
  name: string;
  description: string;
  image_url: string;
  game: Game;
  status: string;
  starts_at: string; // ISO date string
  ends_at: string; // ISO date string
  account_link_url: string;
  self: CampaignSelf;
  time_based_drops: TimeBased[];
  allow: string[];
  deny: string[];
}

// CampaignSelf represents user's relationship to a campaign
export interface CampaignSelf {
  is_account_connected: boolean;
}

// TimeBased represents a time-based drop
export interface TimeBased {
  id: string;
  name: string;
  benefit_edges: BenefitEdge[];
  required_minutes_watched: number;
  self: TimeBasedSelf;
}

// TimeBasedSelf represents user's progress on a time-based drop
export interface TimeBasedSelf {
  current_minutes_watched: number;
  is_claimed: boolean;
  drop_instance_id: string;
}

// BenefitEdge represents a drop benefit
export interface BenefitEdge {
  benefit: Benefit;
}

// Benefit represents a drop reward
export interface Benefit {
  id: string;
  name: string;
  image_url: string;
  is_ios: boolean;
  is_android: boolean;
  game: Game;
}

// PlaybackAccessToken represents stream access token
export interface PlaybackAccessToken {
  value: string;
  signature: string;
}

// WatchingSession represents an active stream watching session
export interface WatchingSession {
  channel_login: string;
  stream_url: string;
  gql_client: any; // GraphQLClient - keeping as any for now
}

// CurrentDropProgress represents current drop progress
export interface CurrentDropProgress {
  current_minutes_watched: number;
  drop_id: string;
}

// GameSlugInfo represents the response from SlugRedirect/DirectoryGameRedirect
export interface GameSlugInfo {
  id: string;
  slug: string;
}

// Inventory represents user's drop inventory
export interface Inventory {
  game_event_drops: GameEventDrop[];
}

// GameEventDrop represents a claimed drop
export interface GameEventDrop {
  id: string;
  benefit: Benefit;
  last_updated_at: string; // ISO date string
  self: DropSelf;
}

// DropSelf represents user's relationship to a drop
export interface DropSelf {
  is_claimed: boolean;
}

// GraphQLResponse represents a GraphQL response
export interface GraphQLResponse {
  data: any;
  errors?: GraphQLError[];
  extensions?: Record<string, any>;
}

// GraphQLError represents a GraphQL error
export interface GraphQLError {
  message: string;
  locations?: GraphQLErrorLocation[];
  path?: any[];
  extensions?: Record<string, any>;
}

// GraphQLErrorLocation represents a GraphQL error location
export interface GraphQLErrorLocation {
  line: number;
  column: number;
}

// TokenResponse represents OAuth token response
export interface TokenResponse {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  scope: string[];
  token_type: string;
}

// ViewerHeartbeatPayload represents WebSocket heartbeat payload
export interface ViewerHeartbeatPayload {
  data: {
    topic: string;
    type: string;
  };
}

// MinerStatus represents the current status of the drop miner
export interface MinerStatus {
  is_running: boolean;
  current_stream?: Stream;
  current_campaign?: Campaign;
  current_progress: number;
  total_campaigns: number;
  claimed_drops: number;
  last_update: string; // ISO date string
  next_switch: string; // ISO date string
  error_message: string;
  active_drops: ActiveDrop[];
}

// ActiveDrop represents a drop that's currently being farmed
export interface ActiveDrop {
  id: string;
  name: string;
  game_name: string;
  required_minutes: number;
  current_minutes: number;
  progress: number;
  is_claimed: boolean;
  estimated_time: string; // ISO date string
}

// MiningSession represents an active mining session
export interface MiningSession {
  id: string;
  user_id: string;
  campaign_id: string;
  stream_id: string;
  started_at: string; // ISO date string
  status: string;
}

// GameConfig represents a game with name, slug, and ID
export interface GameConfig {
  name: string;
  slug: string;
  id: string;
}

// Config represents the application configuration
export interface Config {
  // Server configuration
  server_address: string;
  
  // Twitch API configuration
  twitch_client_id: string;
  
  // Drop mining configuration
  priority_games: GameConfig[];
  claim_drops: boolean;
  webhook_url: string;
  check_interval: number; // seconds
  switch_threshold: number; // minutes
  minimum_points: number;
  maximum_streams: number;
  
  // UI configuration
  theme: string; // "light" or "dark"
  language: string;
  show_tray: boolean;
  start_minimized: boolean;
}

// StoredToken represents a stored OAuth token
export interface StoredToken {
  access_token: string;
  token_type: string;
  expiry: string; // ISO date string
}

// API Response types
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
}

// Auth related types
export interface AuthStatus {
  is_logged_in: boolean;
  user?: User;
}

export interface DeviceCodeResponse {
  device_code: string;
  user_code: string;
  verification_uri: string;
  expires_in: number;
  interval: number;
}

// WebSocket message types
export interface WebSocketMessage {
  type: string;
  data: any;
}

export interface StatusUpdateMessage extends WebSocketMessage {
  type: 'status_update';
  data: MinerStatus;
}

export interface ConfigUpdateMessage extends WebSocketMessage {
  type: 'config_update';
  data: Config;
}

export interface ErrorMessage extends WebSocketMessage {
  type: 'error';
  data: {
    message: string;
  };
}

// Union type for all WebSocket messages
export type WSMessage = StatusUpdateMessage | ConfigUpdateMessage | ErrorMessage;