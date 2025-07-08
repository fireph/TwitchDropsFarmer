# Twitch Drops Farmer (Go + Web)

A complete rewrite of [TwitchDropsMiner](https://github.com/DevilXD/TwitchDropsMiner) in Go with a modern web interface. This application automatically farms Twitch drops by watching streams with the same proven logic as the original Python version.

## Features

- **Automatic Drop Farming**: Watches Twitch streams to earn drops automatically
- **OAuth Authentication**: Secure login via Twitch OAuth flow
- **Real-time Updates**: WebSocket-powered live status updates
- **Modern Web UI**: Dark mode, responsive design with TailwindCSS
- **Campaign Management**: Automatic detection and switching between campaigns
- **Progress Tracking**: Real-time progress monitoring for all active drops
- **Auto-claiming**: Automatically claims completed drops
- **Stream Selection**: Smart stream selection based on viewer count and availability

## Screenshots

Similar layout to the original TwitchDropsMiner but with a modern, dark-mode interface.

## Quick Start

### Prerequisites

- Go 1.21 or higher
- No Twitch app setup required! (Uses official Android app like TDM)

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd TwitchDropsFarmer
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the application:
```bash
go run main.go
```

4. Open your browser and navigate to `http://localhost:8080`

**Note**: All user data is stored in the `./config` directory (login tokens, settings, database).

## Configuration

### Environment Variables

- `TWITCH_CLIENT_ID`: Twitch client ID (defaults to Android app, no change needed)
- `TWITCH_CLIENT_SECRET`: Not required for device flow
- `SERVER_ADDRESS`: Server listen address (default: `:8080`)
- `DATABASE_PATH`: SQLite database file path (default: `drops.db`)
- `WEBHOOK_URL`: Optional webhook URL for notifications

### Settings

The application includes a web-based settings interface where you can configure:

- **Priority Games**: Games to prioritize for drop farming
- **Exclude Games**: Games to exclude from farming
- **Auto-claim**: Automatically claim completed drops
- **Check Interval**: How often to check for updates (seconds)
- **Switch Threshold**: How long to watch a stream before switching (minutes)
- **Theme**: Light or dark mode

## API Documentation

The application provides a REST API for programmatic access:

### Authentication Endpoints
- `GET /api/auth/url` - Get OAuth authorization URL
- `POST /api/auth/callback` - Handle OAuth callback
- `POST /api/auth/logout` - Logout user
- `GET /api/auth/status` - Check authentication status

### Miner Endpoints
- `GET /api/miner/status` - Get current miner status
- `POST /api/miner/start` - Start the drop miner
- `POST /api/miner/stop` - Stop the drop miner

### Campaign Endpoints
- `GET /api/campaigns/` - List all campaigns
- `GET /api/campaigns/:id` - Get specific campaign
- `GET /api/campaigns/:id/drops` - Get drops for campaign

### User Endpoints
- `GET /api/user/profile` - Get user profile
- `GET /api/user/inventory` - Get user's drop inventory

## WebSocket Events

Real-time updates are provided via WebSocket at `/ws`:

- `status_update`: Miner status changes
- `notification`: System notifications
- `error`: Error messages

## Development

### Project Structure

The project follows the guidelines in `CLAUDE.md`:

```
/
├── main.go                 # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── twitch/            # Twitch API client
│   ├── drops/             # Drop mining logic
│   ├── storage/           # Database operations
│   └── web/               # Web server and handlers
├── web/static/            # Frontend assets
│   ├── html/              # HTML templates
│   ├── css/               # Stylesheets
│   └── js/                # JavaScript modules
└── CLAUDE.md              # Development guidelines
```

### Building

```bash
# Build for current platform
go build -o twitchdropsminer

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o twitchdropsminer-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o twitchdropsminer.exe

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o twitchdropsminer-macos
```

### Running in Production

1. Set `GIN_MODE=release` in your environment
2. Use a reverse proxy (nginx/Apache) for HTTPS
3. Set up proper firewall rules
4. Consider using a process manager like systemd

## Security Considerations

- OAuth tokens are stored securely in SQLite database
- CORS is configured for web interface
- Rate limiting is implemented for API endpoints
- No sensitive data is logged

## Contributing

1. Follow the architecture guidelines in `CLAUDE.md`
2. Maintain compatibility with original TwitchDropsMiner logic
3. Use Go best practices and error handling
4. Test thoroughly before submitting PRs

## License

This project maintains the same license as the original TwitchDropsMiner.

## Acknowledgments

- [DevilXD](https://github.com/DevilXD) for the original TwitchDropsMiner
- Twitch for providing the API infrastructure
- The Go and web development communities

## Disclaimer

This tool is for educational purposes. Use responsibly and in accordance with Twitch's Terms of Service.