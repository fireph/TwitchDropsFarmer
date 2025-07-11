# TwitchDropsFarmer - Claude Code Organization Guide

## Project Overview
This is a Go backend + Vue.js 3 + TypeScript + TailwindCSS frontend implementation of TwitchDropsFarmer, designed to automatically farm Twitch drops with minimal user intervention.

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
    graphql_client.go # GraphQL client implementation
    operations.go    # GraphQL operations and queries
    types.go         # Twitch API data structures
  /drops/            # Drop mining logic
    miner.go         # Core drop mining engine
  /config/           # Configuration management
    settings.go      # Application settings and config loading
  /web/              # Web server and API
    server.go        # HTTP server setup
    handlers.go      # HTTP request handlers
    middleware.go    # HTTP middleware
  /util/             # Shared utilities
    util.go          # General utilities
```

#### Code Organization Principles:
- **Single Responsibility**: Each package should have one clear purpose
- **Minimal Dependencies**: Keep internal package dependencies minimal
- **Interface Segregation**: Define small, focused interfaces
- **Error Handling**: Use Go's explicit error handling consistently
- **Context Propagation**: Use context.Context for cancellation and timeouts

### Frontend (Vue.js 3 + TypeScript + TailwindCSS)
Organize frontend code into logical, maintainable components:

#### Frontend Structure:
```
/web/
  /src/
    /components/     # Reusable UI components
      CampaignCard.vue # Campaign display component
      DropCard.vue     # Individual drop component
      SettingsCard.vue # Settings UI component
      StatusCard.vue   # Status display component
      StreamCard.vue   # Stream information component
    /views/          # Page-level components
      Dashboard.vue    # Main dashboard view
      Login.vue        # Authentication view
    /stores/         # Pinia state management
      auth.ts          # Authentication state
      miner.ts         # Drop mining state
      theme.ts         # Theme management
    /services/       # API interaction services
      api.ts           # Base API client
      websocket.ts     # WebSocket client
    /types/          # TypeScript type definitions
      index.ts         # All type definitions
    /router/         # Vue Router configuration
      index.ts         # Route definitions
    /assets/         # Static assets
    /composables/    # Vue composables (reusable logic)
    App.vue          # Root application component
    main.ts          # Application entry point
    style.css        # Global styles
  vite.config.ts     # Vite configuration
  package.json       # Dependencies and scripts
  tsconfig.json      # TypeScript configuration
```

#### Vue.js Component Organization Standards:

##### Component Structure:
- **Single File Components**: Use .vue files with `<template>`, `<script setup>`, and `<style>` sections
- **Composition API**: Use `<script setup>` with Composition API for better TypeScript support
- **Props Interface**: Define TypeScript interfaces for component props
- **Emit Events**: Use `defineEmits()` for component events
- **Reactive State**: Use `ref()` and `reactive()` appropriately

##### Component Naming:
- **PascalCase**: All component files use PascalCase (e.g., `CampaignCard.vue`)
- **Descriptive Names**: Component names should clearly indicate their purpose
- **Suffix Convention**: Use descriptive suffixes like Card, Modal, Form, List

##### Component Responsibilities:
- **Single Purpose**: Each component should have one clear responsibility
- **Reusability**: Design components to be reusable across different contexts
- **Props vs Slots**: Use props for data, slots for content injection
- **Event Handling**: Emit events for parent communication, avoid direct prop mutation

##### State Management:
- **Pinia Stores**: Use Pinia for global state management
- **Store Organization**: Separate stores by domain (auth, miner, theme)
- **Composition API**: Use Pinia's Composition API syntax
- **Computed Properties**: Use computed properties for derived state

##### Styling Guidelines:
- **TailwindCSS**: Use utility classes for styling
- **Dark Mode**: Support dark mode with `dark:` prefixes
- **Responsive Design**: Use responsive utility classes
- **Component Scoping**: Use scoped styles when needed for component-specific CSS

#### Frontend Coding Standards:

##### TypeScript Standards:
- **Strict Types**: Use strict TypeScript configuration
- **Interface Definitions**: Define interfaces for all data structures
- **Type Safety**: Avoid `any` type unless absolutely necessary
- **Generic Types**: Use generics for reusable components and functions

##### Vue 3 Best Practices:
- **Composition API**: Prefer Composition API over Options API
- **Script Setup**: Use `<script setup>` for concise component definition
- **Reactivity**: Use `ref()` for primitives, `reactive()` for objects
- **Lifecycle Hooks**: Use composition API lifecycle hooks
- **Template Refs**: Use `ref()` for DOM element references

##### Performance Considerations:
- **Lazy Loading**: Implement lazy loading for routes and components
- **Tree Shaking**: Ensure dead code elimination
- **Bundle Optimization**: Use Vite's built-in optimizations
- **Image Optimization**: Handle image loading errors gracefully

## Key Implementation Requirements

### Authentication Flow
- Implement OAuth device flow matching TwitchDropsMiner's approach
- Support 2FA and token refresh handling
- Maintain session persistence across browser refreshes
- Handle authentication state in Pinia store

### Drop Mining Logic
- **Campaign Discovery**: Fetch and filter available campaigns
- **Progress Tracking**: Real-time drop progress via WebSocket
- **Automatic Management**: Handle drop claiming and campaign switching
- **State Synchronization**: Keep frontend state in sync with backend

### Web Interface Features
- **Responsive Dashboard**: Real-time status of mining operations
- **Authentication Flow**: OAuth login with device verification
- **Settings Management**: Configurable preferences and game exclusions
- **Progress Visualization**: Display of campaign and drop progress
- **Real-time Updates**: WebSocket-based live updates

## Component Design Patterns

### Card Components
- Use consistent card layout patterns
- Support both light and dark themes
- Include loading and error states
- Implement proper accessibility features

### Form Components
- Use TypeScript for form validation
- Implement proper error handling
- Support reactive form state
- Include loading states for async operations

### List Components
- Support virtualization for large lists
- Implement proper loading states
- Handle empty states gracefully
- Use consistent item spacing and layout

## File Organization Rules

1. **Keep components focused**: Each component should have a single, clear responsibility
2. **Use descriptive names**: Component names should clearly indicate their purpose
3. **Group related functionality**: Related components should be in the same directory
4. **Separate concerns**: UI logic separate from business logic separate from data access
5. **Minimal component size**: Aim for components under 300 lines; split larger components

## Development Workflow

### When adding new features:
1. Start with defining types in `types/index.ts`
2. Update Pinia stores if global state is needed
3. Create/update Vue components following the established patterns
4. Add API endpoints in Go backend
5. Update WebSocket handlers if real-time updates are needed
6. Test component functionality and responsiveness

### Component Development Process:
1. **Plan Component Structure**: Define props, emits, and internal state
2. **Create TypeScript Interfaces**: Define all data structures
3. **Implement Template**: Create semantic HTML with TailwindCSS
4. **Add Logic**: Implement component behavior with Composition API
5. **Handle States**: Add loading, error, and empty states
6. **Test Responsiveness**: Ensure component works on all screen sizes

## Code Quality Standards

### Vue.js Standards:
- **ESLint**: Use Vue-specific ESLint rules
- **Vetur/Volar**: Use appropriate Vue language server
- **TypeScript**: Strict type checking enabled
- **Accessibility**: Ensure ARIA attributes and semantic HTML
- **Performance**: Optimize re-renders and memory usage

### Testing Strategy:
- **Unit Tests**: Test individual components in isolation
- **Integration Tests**: Test component interactions
- **E2E Tests**: Test complete user workflows
- **Mock Services**: Mock API calls and WebSocket connections

## Dependencies Guidelines

### Go Dependencies:
- Use standard library when possible
- Prefer well-maintained, minimal dependencies
- Pin dependency versions in go.mod
- Regular security updates

### Frontend Dependencies:
- **Vue.js 3**: Latest stable version with Composition API
- **TypeScript**: For type safety and better developer experience
- **Pinia**: For state management (Vue 3 recommended store)
- **Vue Router**: For client-side routing
- **TailwindCSS**: For utility-first styling
- **Vite**: For fast development and optimized builds

### Development Tools:
- **Vite**: Fast development server and build tool
- **Vue TSC**: TypeScript compiler for Vue files
- **ESLint**: Code linting with Vue-specific rules
- **Prettier**: Code formatting (if needed)

## Performance Optimization

### Frontend Performance:
- **Code Splitting**: Implement route-based code splitting
- **Lazy Loading**: Lazy load components and routes
- **Tree Shaking**: Ensure unused code is eliminated
- **Bundle Analysis**: Monitor bundle size and optimize imports
- **Image Optimization**: Handle image loading and errors properly

### Vue-Specific Optimizations:
- **v-memo**: Use for expensive list rendering
- **KeepAlive**: Cache components when appropriate
- **Async Components**: Load components asynchronously
- **Virtual Scrolling**: For large lists (if needed)

This structure ensures maintainable, scalable Vue.js components that closely match modern frontend development best practices while maintaining high code quality and performance standards.