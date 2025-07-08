# TwitchDropsFarmer - Claude Code Organization Guide

## Project Overview
This is a Go backend + TypeScript/HTML/TailwindCSS frontend implementation of TwitchDropsMiner, designed to automatically farm Twitch drops with minimal user intervention.

## Architecture Guidelines

### Backend (Go)
Organize Go code into clear, single-responsibility packages:

#### Core Packages Structure:
```
main.go              # Main application entry point
/internal/
  /twitch/           # Twitch API interactions
    client.go        # Main Twitch client
    auth.go          # OAuth authentication flow
    graphql.go       # GraphQL operations
    websocket.go     # WebSocket connections
    types.go         # Twitch API data structures
  /drops/            # Drop mining logic
    miner.go         # Core drop mining engine
    campaign.go      # Campaign management
    inventory.go     # User inventory tracking
    progress.go      # Drop progress tracking
  /channels/         # Channel management
    manager.go       # Channel switching logic
    selector.go      # Channel selection algorithms
    watcher.go       # Stream watching coordination
  /config/           # Configuration management
    settings.go      # Application settings
    loader.go        # Config file loading
    validation.go    # Config validation
  /web/              # Web server and API
    server.go        # HTTP server setup
    handlers.go      # HTTP request handlers
    middleware.go    # HTTP middleware
    routes.go        # Route definitions
  /storage/          # Data persistence
    database.go      # Database operations
    models.go        # Data models
    migrations.go    # Database migrations
  /utils/            # Shared utilities
    logging.go       # Logging utilities
    errors.go        # Error handling
    http.go          # HTTP utilities
```

#### Code Organization Principles:
- **Single Responsibility**: Each package should have one clear purpose
- **Minimal Dependencies**: Keep internal package dependencies minimal
- **Interface Segregation**: Define small, focused interfaces
- **Error Handling**: Use Go's explicit error handling consistently
- **Context Propagation**: Use context.Context for cancellation and timeouts

### Frontend (TypeScript + HTML + TailwindCSS)
Organize frontend code into logical components:

#### Frontend Structure:
```
/web/
  /static/
    /js/
      /components/     # Reusable UI components
        auth.ts        # Authentication UI
        dashboard.ts   # Main dashboard
        inventory.ts   # Inventory display
        settings.ts    # Settings management
        channels.ts    # Channel management UI
      /services/       # API interaction services
        api.ts         # Base API client
        twitch.ts      # Twitch-specific API calls
        websocket.ts   # WebSocket client
      /utils/          # Frontend utilities
        dom.ts         # DOM manipulation helpers
        storage.ts     # Local storage utilities
        formatting.ts  # Data formatting utilities
      /types/          # TypeScript type definitions
        api.ts         # API response types
        twitch.ts      # Twitch data types
        ui.ts          # UI component types
      main.ts          # Application entry point
    /css/
      main.css         # Main styles with TailwindCSS
      components.css   # Component-specific styles
    /html/
      index.html       # Main application page
      login.html       # Authentication page
```

#### Frontend Coding Standards:
- **Component-Based**: Create reusable UI components
- **Type Safety**: Use TypeScript interfaces for all data structures
- **Async/Await**: Use modern async patterns for API calls
- **Responsive Design**: Use TailwindCSS classes for responsive layouts
- **Error Boundaries**: Implement proper error handling in UI components

## Key Implementation Requirements

### Authentication Flow
- Implement OAuth device flow matching TwitchDropsMiner's approach
- Support 2FA and captcha handling
- Maintain session persistence
- Handle token refresh automatically

### Drop Mining Logic
- **Campaign Discovery**: Fetch and filter available campaigns
- **Channel Selection**: Implement priority-based channel selection
- **Progress Tracking**: Real-time drop progress via WebSocket
- **Automatic Switching**: Switch channels when drops complete
- **Inventory Management**: Track claimed drops and rewards

### Web Interface Features
- **Dashboard**: Real-time status of mining operations
- **Authentication**: OAuth login flow with device verification
- **Settings**: Configurable preferences and game exclusions
- **Inventory**: Display of claimed drops and progress
- **Logs**: Real-time logging display

## File Organization Rules

1. **Keep files focused**: Each file should have a single, clear responsibility
2. **Use descriptive names**: File names should clearly indicate their purpose
3. **Group related functionality**: Related types and functions should be in the same file
4. **Separate concerns**: UI logic separate from business logic separate from data access
5. **Minimal file size**: Aim for files under 300 lines; split larger files

## Development Workflow

### When adding new features:
1. Start with defining types in appropriate `types.go` or `types.ts` files
2. Implement core logic in business layer (Go packages)
3. Add API endpoints in web handlers
4. Create frontend components and services
5. Add appropriate error handling at all layers
6. Update configuration if needed

### Testing Strategy:
- Unit tests for all business logic
- Integration tests for API endpoints
- End-to-end tests for critical user flows
- Mock external dependencies (Twitch API)

## Code Quality Standards

- **Linting**: Use golangci-lint for Go, ESLint for TypeScript
- **Formatting**: Use gofmt for Go, prettier for frontend
- **Documentation**: Document public APIs and complex logic
- **Error Messages**: Provide clear, actionable error messages
- **Performance**: Optimize for minimal resource usage

## Dependencies Guidelines

### Go Dependencies:
- Use standard library when possible
- Prefer well-maintained, minimal dependencies
- Pin dependency versions
- Regular security updates

### Frontend Dependencies:
- Minimize JavaScript bundle size
- Use TailwindCSS for styling
- Prefer vanilla TypeScript over heavy frameworks
- Ensure accessibility compliance

This structure ensures maintainable, scalable code that closely matches the original TwitchDropsMiner functionality while leveraging modern Go and web technologies.