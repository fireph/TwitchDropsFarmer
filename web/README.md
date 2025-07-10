# Twitch Drops Farmer - Frontend

This is the modern Vue.js frontend for the Twitch Drops Farmer application.

## Tech Stack

- **Vue.js 3.5** - Progressive JavaScript framework
- **TypeScript 5.7** - Type-safe JavaScript
- **Vite 6.0** - Fast build tool and development server
- **TailwindCSS v4** - Utility-first CSS framework with CSS-based configuration
- **Vue Router 4.4** - Client-side routing
- **Pinia 2.2** - State management
- **WebSocket** - Real-time updates from Go backend

## Development

### Prerequisites

- Node.js 18 or higher
- npm

### Setup

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build
```

## Architecture

### Project Structure

```
web/
├── src/
│   ├── components/        # Reusable Vue components
│   ├── views/            # Route pages (Dashboard, Login)
│   ├── stores/           # Pinia stores for state management
│   ├── services/         # API and WebSocket services
│   ├── types/            # TypeScript type definitions
│   ├── router/           # Vue Router configuration
│   └── main.ts           # Application entry point
├── public/               # Static assets
├── static/               # Built files (generated)
└── package.json
```

### Key Features

1. **TypeScript Types** - Exact matches for all Go structs
2. **Real-time Updates** - WebSocket connection for live status
3. **Responsive Design** - Works on desktop and mobile
4. **Dark Mode** - Automatic theme switching
5. **Vue Router** - Single page application routing
6. **State Management** - Pinia stores for auth, miner, theme
7. **Modern TailwindCSS v4** - CSS-based configuration with custom theme variables

### API Integration

The frontend communicates with the Go backend through:
- REST API endpoints for actions (start/stop miner, save config)
- WebSocket connection for real-time status updates
- Same authentication flow as the original implementation

## Build Process

The build process is integrated into the main `start.sh` script:

1. Install npm dependencies
2. Build Vue.js application with Vite
3. Output to `web/static/` directory
4. Go server serves the built files

## Authentication Flow

The login flow maintains the same OAuth device flow as the original:

1. User clicks "Login with Twitch"
2. Application requests device code from backend
3. User enters code on Twitch activation page
4. Frontend polls backend for authentication status
5. Redirects to dashboard when authenticated

All authentication is handled by Vue Router with navigation guards.

## TailwindCSS v4 Configuration

TailwindCSS v4 uses CSS-based configuration instead of a separate config file. Custom theme variables are defined in `src/style.css`:

```css
@theme {
  --color-twitch-purple: #9146FF;
  --color-twitch-purple-dark: #772CE8;
}
```

These can be used in components as:
- `bg-twitch-purple` - Twitch brand purple
- `bg-twitch-purple-dark` - Darker variant for hover states
- `text-twitch-purple` - Purple text color

No separate `tailwind.config.js` file is needed.