# Media Manager Web Frontend

Modern React-based frontend for the Media Manager application.

## Tech Stack

- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State Management**: Zustand + React Query
- **Routing**: React Router
- **Icons**: Lucide React

## Project Structure

```
web/
├── src/
│   ├── components/       # Reusable UI components
│   │   ├── ui/          # Base UI components (Button, Card, etc.)
│   │   ├── Layout.tsx   # Main layout wrapper
│   │   ├── Header.tsx   # Top navigation bar
│   │   ├── Sidebar.tsx  # Side navigation
│   │   └── Toaster.tsx  # Toast notifications
│   ├── pages/           # Page components
│   │   ├── Dashboard.tsx
│   │   ├── Downloads.tsx
│   │   ├── Library.tsx
│   │   ├── Search.tsx
│   │   ├── Suggestions.tsx
│   │   └── Settings.tsx
│   ├── hooks/           # Custom React hooks
│   │   └── useApi.ts    # API hooks with React Query
│   ├── lib/             # Utility functions
│   │   ├── api.ts       # API client
│   │   └── utils.ts     # Helper functions
│   ├── store/           # State management
│   │   └── appStore.ts  # Zustand store
│   ├── types/           # TypeScript types
│   │   └── index.ts     # API types matching Go models
│   ├── App.tsx          # Main app component
│   ├── main.tsx         # Entry point
│   └── index.css        # Global styles
├── public/              # Static assets
├── index.html           # HTML entry point
├── package.json         # Dependencies
├── vite.config.ts       # Vite configuration
├── tailwind.config.js   # Tailwind CSS configuration
└── tsconfig.json        # TypeScript configuration
```

## Development

### Prerequisites

- Node.js 18+ and npm
- Backend server running (Go service on port 8081)

### Setup

```bash
cd web
npm install
```

### Run Development Server

```bash
npm run dev
```

This starts the Vite dev server on port 5173 with hot reload and API proxy to the backend.

### Build for Production

```bash
npm run build
```

This creates an optimized production build in `web/dist/`.

## Architecture

### API Integration

The frontend uses a layered API approach:

1. **API Client** (`lib/api.ts`): Axios-based HTTP client with typed endpoints
2. **React Query Hooks** (`hooks/useApi.ts`): Data fetching with caching, refetching, and mutations
3. **Components**: Use hooks for data and mutations

### State Management

- **Zustand** (`store/appStore.ts`): Global UI state (theme, sidebar, toasts, selections)
- **React Query**: Server state (API data) with automatic caching and synchronization

### Routing

React Router handles client-side routing:

- `/` - Dashboard
- `/downloads` - Download queue
- `/library` - Media library
- `/search` - Torrent search
- `/suggestions` - AI suggestions
- `/settings` - Configuration

The backend serves the SPA catch-all handler for all routes.

## AI-Friendly Structure

This codebase is optimized for AI code generation:

1. **Clear Types**: TypeScript interfaces mirror Go structs exactly
2. **Component Hierarchy**: Consistent structure with clear props
3. **API Layer**: Well-defined endpoints with typed responses
4. **Utility Functions**: Reusable helpers for formatting and styling
5. **Consistent Patterns**: Follows React best practices throughout

## API Endpoints

The frontend communicates with the Go backend via these endpoints:

- `GET /api/stats` - Download statistics
- `GET /api/tasks` - Download tasks list
- `POST /api/tasks/cancel` - Cancel a task
- `POST /api/tasks/restart` - Restart a task
- `GET /api/suggestions` - AI suggestions
- `POST /api/suggestions/generate` - Generate new suggestions
- `GET /api/library/movies` - Media library
- `GET /api/sources` - Download sources
- `GET /api/rules` - Download rules

All endpoints return JSON and support CORS for development.

## Customization

### Themes

The app supports light/dark modes via CSS variables. Toggle in the Header component.

### Styling

Uses Tailwind CSS utility classes. Customize in `tailwind.config.js`.

### Icons

Lucide React icons. Replace with any icon library by updating imports.

## Building with Go

The Go backend serves the built React files:

1. Build the frontend: `npm run build`
2. Build the Go service: `go build ./cmd/media-manager-service`
3. The Go server serves static files from `web/dist/`

See the root Makefile for combined build commands.
