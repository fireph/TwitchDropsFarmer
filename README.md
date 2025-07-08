# TwitchDropsFarmer

A Go-based implementation of TwitchDropsMiner with a modern web interface and comprehensive API. This application automatically farms Twitch drops by watching streams using the exact same proven methods as the original TwitchDropsMiner.

## Key Features

- **🎯 Exact TDM Logic**: Uses identical GraphQL operations and authentication methods as TwitchDropsMiner
- **📊 Real-time Progress**: Live drop progress tracking using Twitch's DropCurrentSessionContext API
- **🔄 Sequential Drop Support**: Handles multi-drop campaigns (30min → 90min → 180min sequences)
- **🌐 Modern Web Interface**: Real-time dashboard with WebSocket updates
- **🔧 Comprehensive API**: Full REST API for external integrations
- **🤖 Device Flow Auth**: Secure OAuth using Twitch's Android app credentials (no app setup required)
- **⚡ Auto-switching**: Intelligent stream selection and campaign switching
- **📱 Mobile-friendly**: Responsive design that works on all devices

## Screenshots

Similar layout to the original TwitchDropsMiner but with a modern, dark-mode interface.

![TwitchDropsFarmer Web UI](https://github.com/user-attachments/assets/c90f7eda-557b-47ad-9dfe-145f1923cd57)

## Quick Start

### Prerequisites

- Go 1.24 or higher
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
go run .
```

4. Open your browser and navigate to `http://localhost:8080`

**Note**: All user data is stored locally (tokens in SQLite, settings in config files).

## Configuration

### Environment Variables

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

The application provides a comprehensive REST API for programmatic access:

### Authentication Endpoints
- `GET /api/auth/url` - Get OAuth device flow authorization URL
- `POST /api/auth/callback` - Complete OAuth device flow with device code
- `POST /api/auth/logout` - Logout and revoke tokens
- `GET /api/auth/status` - Check authentication status and user info

### Drop Mining Endpoints
- `GET /api/miner/status` - Get detailed miner status (campaigns, streams, progress)
- `GET /api/miner/current-drop` - Get currently active drop with real-time progress
- `GET /api/miner/progress` - Get progress for all drops (completed + current + pending)
- `POST /api/miner/start` - Start the drop mining process
- `POST /api/miner/stop` - Stop the drop mining process

### Campaign Endpoints
- `GET /api/campaigns/` - List all available drop campaigns
- `GET /api/campaigns/:id` - Get detailed campaign information
- `GET /api/campaigns/:id/drops` - Get all drops for a specific campaign

### User Endpoints
- `GET /api/user/profile` - Get authenticated user profile
- `GET /api/user/inventory` - Get user's claimed drops inventory

### Settings Endpoints
- `GET /api/settings` - Get current application settings
- `PUT /api/settings` - Update application settings

### Stream Endpoints
- `GET /api/streams/game/:gameId?limit=10` - Get live streams for a specific game
- `GET /api/streams/current` - Get currently watched stream

### Example API Usage

```bash
# Check if miner is running and what it's doing
curl http://localhost:8080/api/miner/status

# Get real-time progress for all drops
curl http://localhost:8080/api/miner/progress

# Get currently active drop with live progress
curl http://localhost:8080/api/miner/current-drop

# Start drop mining
curl -X POST http://localhost:8080/api/miner/start

# Get all available campaigns
curl http://localhost:8080/api/campaigns/
```

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
go build -o twitchdropsfarmer

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o twitchdropsfarmer-linux

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o twitchdropsfarmer.exe

# Build for macOS
GOOS=darwin GOARCH=amd64 go build -o twitchdropsfarmer-macos
```

### Live Development

For development with auto-reload:

```bash
# Install air for live reloading
go install github.com/air-verse/air@latest

# Run with auto-reload
air
```

### Running in Production

1. Set `GIN_MODE=release` in your environment
2. Use a reverse proxy (nginx/Apache) for HTTPS
3. Set up proper firewall rules
4. Consider using a process manager like systemd

## Technical Implementation

### Drop Progress Tracking

This implementation uses the exact same approach as TwitchDropsMiner:

1. **Real-time Progress**: Uses Twitch's `DropCurrentSessionContext` GraphQL operation to get live progress data that matches exactly what appears on twitch.tv
2. **Sequential Drop Logic**: For multi-drop campaigns (e.g., 30min → 90min → 180min), automatically determines completion status of previous drops based on the currently active drop
3. **Accurate Channel Targeting**: Uses the correct channel user ID (not stream ID) for GraphQL operations

### Authentication

- Uses Twitch's OAuth Device Flow with Android app credentials (same as TDM)
- No need to create your own Twitch app
- Tokens are stored securely and refreshed automatically

## Security Considerations

- OAuth tokens are stored securely in SQLite database
- CORS is configured for web interface  
- No sensitive data is logged
- Uses HTTPS-ready configuration

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
