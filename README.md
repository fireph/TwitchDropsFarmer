# TwitchDropsFarmer

A modern, web-based application for automatically farming Twitch drops with a beautiful SolidJS frontend and Go backend. Designed to replicate and enhance the functionality of [DevilXD's TwitchDropsMiner](https://github.com/DevilXD/TwitchDropsMiner) with a modern web interface.

## ✨ Features

- **🎮 Game Management**: Add games by name to your watch list with drag-and-drop priority ordering
- **📱 Modern Web UI**: Beautiful, responsive interface built with SolidJS and TailwindCSS
- **🔄 Real-time Updates**: Live status updates and logging via WebSocket connection
- **🔐 SmartTV Login**: Uses the same Twitch SmartTV OAuth flow as TwitchDropsMiner for seamless authentication
- **📊 Drop Tracking**: View all available drops for your games with progress tracking
- **⚙️ Configurable Settings**: Customize watch intervals, auto-claim drops, and appearance
- **🐳 Docker Ready**: Easy deployment with Docker and docker-compose
- **🌙 Dark Mode**: Beautiful dark theme optimized for extended use
- **📱 Mobile Responsive**: Works perfectly on desktop, tablet, and mobile devices

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/yourusername/TwitchDropsFarmer.git
cd TwitchDropsFarmer
```

2. Start the application:
```bash
docker-compose up -d
```

3. Open your browser and navigate to `http://localhost:8080`

4. Click "Login with Twitch" and follow the SmartTV authentication process

5. Add games to your watch list and start farming!

### Manual Installation

#### Prerequisites
- Go 1.21+
- Node.js 18+
- npm or yarn

#### Backend Setup
```bash
# Install Go dependencies
go mod download

# Build the backend
go build -o twitchdropsfarmer ./cmd/server
```

#### Frontend Setup
```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Build the frontend
npm run build
```

#### Run the Application
```bash
# Set environment variables
export PORT=8080
export DATA_DIR=./data
export ENVIRONMENT=production

# Run the application
./twitchdropsfarmer
```

## 🎯 How It Works

TwitchDropsFarmer works exactly like DevilXD's TwitchDropsMiner but with a modern web interface:

1. **Stream-less Mining**: Uses GraphQL queries to simulate watching streams without downloading video data
2. **Smart Channel Switching**: Automatically switches between channels when streams go offline
3. **Priority-based Gaming**: Games are watched in the order you specify in your watch list
4. **Drop Detection**: Automatically detects available drops for your linked accounts
5. **Real-time Logging**: See exactly what's happening with detailed, real-time logs

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Port to run the web server on |
| `DATA_DIR` | `/data` | Directory for persistent data storage |
| `ENVIRONMENT` | `development` | Application environment (`development` or `production`) |
| `TWITCH_CLIENT_ID` | `kimne78kx3ncx6brgo4mv6wki5h1ko` | Twitch client ID (uses same as TwitchDropsMiner) |
| `TWITCH_CLIENT_SECRET` | - | Twitch client secret (optional for SmartTV flow) |

### Application Settings

Access the Settings tab in the web interface to configure:

- **Watch Interval**: How often to send watch requests (10-120 seconds)
- **Auto-claim Drops**: Automatically claim completed drops
- **Notifications**: Enable browser notifications for important events
- **Theme**: Choose between light, dark, or auto themes
- **Language**: Select your preferred language

## 📋 API Compatibility

TwitchDropsFarmer uses the **exact same** GraphQL queries and client IDs as TwitchDropsMiner:

- **Client ID**: `kimne78kx3ncx6brgo4mv6wki5h1ko` (Twitch's web client ID)
- **GraphQL Endpoint**: `https://gql.twitch.tv/gql`
- **Authentication**: SmartTV OAuth2 device flow
- **Watch Requests**: Identical timing and methodology

## 🗂️ Project Structure

```
TwitchDropsFarmer/
├── cmd/server/          # Go application entry point
├── internal/
│   ├── api/            # HTTP API handlers
│   ├── config/         # Configuration management
│   ├── miner/          # Drop mining logic
│   ├── storage/        # Data persistence
│   └── twitch/         # Twitch API client
├── frontend/
│   ├── src/
│   │   ├── components/ # SolidJS components
│   │   ├── context/    # React-style contexts
│   │   └── App.tsx     # Main application component
│   ├── package.json
│   └── vite.config.ts
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## 🛠️ Development

### Frontend Development
```bash
cd frontend
npm run dev
```

### Backend Development
```bash
# Run with auto-reload (install air first: go install github.com/cosmtrek/air@latest)
air

# Or run manually
go run ./cmd/server
```

### Building for Production
```bash
# Build frontend
cd frontend && npm run build && cd ..

# Build backend
go build -o twitchdropsfarmer ./cmd/server

# Or use Docker
docker build -t twitchdropsfarmer .
```

## 🔐 Security & Privacy

- **No Password Storage**: Uses OAuth2 device flow, no passwords stored
- **Local Data**: All data stored locally in Docker volumes or specified data directory
- **Secure Headers**: Production builds include security headers
- **No Telemetry**: No data collection or external reporting

## 🤝 Compatibility with TwitchDropsMiner

TwitchDropsFarmer is designed to be a drop-in replacement for TwitchDropsMiner:

- ✅ Same authentication method (SmartTV OAuth)
- ✅ Same Twitch client ID and GraphQL queries
- ✅ Same watch timing and intervals
- ✅ Compatible with all drop campaigns
- ✅ Same stream switching logic
- ✅ Identical API interactions with Twitch

## 📊 Monitoring & Logs

- **Real-time Logs**: View mining activity in the web interface
- **Health Checks**: Built-in health check endpoint at `/health`
- **WebSocket Updates**: Live status and log updates
- **Docker Logs**: Standard Docker logging for container monitoring

## 🚨 Important Notes

- **Account Linking**: Make sure your Twitch account is linked to game accounts on the [Twitch Drops page](https://www.twitch.tv/drops/campaigns)
- **One Instance**: Only run one instance per Twitch account to avoid conflicts
- **Browser Watching**: Don't watch streams in your browser while farming (can cause issues)
- **Rate Limiting**: Respects Twitch's rate limits and uses appropriate intervals

## 📄 License

This project is open source and available under the [MIT License](LICENSE).

## 🙏 Acknowledgments

- **DevilXD** for the original [TwitchDropsMiner](https://github.com/DevilXD/TwitchDropsMiner) that inspired this project
- **Twitch** for providing the GraphQL API
- **SolidJS** and **Go** communities for excellent frameworks

## 🐛 Issues & Support

If you encounter any issues:

1. Check the logs in the web interface
2. Verify your Twitch account is properly linked
3. Ensure no other drop farmers are running
4. Create an issue on GitHub with detailed information

## 🔄 Updates

TwitchDropsFarmer will be updated to maintain compatibility with Twitch's API changes and include new features based on community feedback.

---

**Happy Drop Farming! 🎮✨**