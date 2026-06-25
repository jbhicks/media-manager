# Media Manager - Entertainment Center Upgrade

## 🎬 Overview

This upgrade transforms the media-manager from a torrent download tool into a fully-featured entertainment discovery and management platform with a Spotify-inspired dark UI.

## ✅ Completed Features

### Phase 1: Foundation

#### 1. Authentication System
- **File**: `internal/service/auth.go`
- **Features**:
  - JWT-based authentication with middleware
  - User registration and login endpoints
  - Password hashing with SHA-256
  - Token-based session management (24h expiry)
  - User roles (admin, user)
  - Protected route middleware

#### 2. TV Show Models
- **File**: `pkg/models/tv_models.go`
- **Models**:
  - `TVShow` - Series metadata with TMDb integration
  - `Season` - Individual seasons with episode counts
  - `Episode` - Episode details with runtime, ratings, air dates
  - `CastMember` - Actor/actress information
  - `CrewMember` - Director, writer, producer info
  - `Genre` - Movie/TV genre categorization
  - `Video` - Trailer and clip metadata

#### 3. User Models
- **File**: `pkg/models/user_models.go`
- **Models**:
  - `User` - Authentication and profile
  - `WatchHistory` - Track viewing progress and resume points
  - `Watchlist` - Save movies/shows for later
  - `UserPreference` - Quality, codec, subtitle preferences

#### 4. TMDB Discover API
- **File**: `internal/service/discover.go`
- **Endpoints**:
  - `GET /api/discover/movies/trending` - Trending movies this week
  - `GET /api/discover/movies/popular` - Popular movies
  - `GET /api/discover/movies/now_playing` - In theaters now
  - `GET /api/discover/movies/upcoming` - Coming soon
  - `GET /api/discover/movies/top_rated` - Highest rated
  - `GET /api/discover/movies/by_genre?genre_id=X` - Genre filter
  - `GET /api/discover/tv/trending` - Trending TV shows
  - `GET /api/discover/tv/popular` - Popular TV shows
  - `GET /api/discover/tv/on_the_air` - Currently airing
  - `GET /api/discover/tv/airing_today` - Airing today
  - `GET /api/discover/tv/top_rated` - Highest rated TV
  - `GET /api/discover/tv/by_genre?genre_id=X` - TV genre filter
  - `GET /api/discover/genres?type=movie|tv` - Genre list
  - `GET /api/discover/movie/:id` - Movie details with cast, crew, similar, videos
  - `GET /api/discover/tv/:id` - TV show details with seasons, episodes, cast
  - `GET /api/discover/tv/:id/season/:num` - Season episodes

### Phase 2: Content Pages

#### 5. Discover Page
- **File**: `web/src/pages/Discover.tsx`
- **Features**:
  - Spotify-inspired dark design (#121212 background)
  - Hero section with backdrop image
  - Tab navigation (All / Movies / TV Shows)
  - Horizontal scrolling content rows
  - Movie/TV cards with poster, title, rating, year
  - Hover effects with scale and rating overlay
  - Animated transitions between tabs
  - Loading skeletons

#### 6. Movie Detail Page
- **File**: `web/src/pages/MovieDetail.tsx`
- **Features**:
  - Full-width backdrop hero
  - Movie title, tagline, rating, year, runtime
  - Genre pills
  - Watch trailer, Add to watchlist, Download buttons
  - Overview section
  - Cast carousel with photos and character names
  - Crew info (directors, writers)
  - Movie info sidebar (budget, revenue, status)
  - Similar movies carousel
  - YouTube trailer modal
  - Back navigation

#### 7. TV Detail Page
- **File**: `web/src/pages/TVDetail.tsx`
- **Features**:
  - Full-width backdrop hero
  - Show title, rating, seasons, episodes
  - Genre pills
  - Season selector dropdown
  - Episode list with thumbnails, descriptions, ratings
  - Play button on episode hover
  - Cast carousel
  - Show info sidebar (network, creator, status)
  - Similar shows carousel
  - YouTube trailer modal

#### 8. Login Page
- **File**: `web/src/pages/Login.tsx`
- **Features**:
  - Dark themed login form
  - Username/password fields
  - Toggle between login/register
  - Loading states
  - Error messages
  - JWT token storage in localStorage

### Phase 3: UI/UX

#### 9. Auth Context
- **File**: `web/src/contexts/AuthContext.tsx`
- **Features**:
  - React context for auth state
  - Login/register/logout functions
  - Token persistence in localStorage
  - User info storage

#### 10. Navigation Updates
- **File**: `web/src/components/Sidebar.tsx`
- **Changes**:
  - Added "Discover" link to sidebar
  - Compass icon for discovery

#### 11. Routing
- **File**: `web/src/App.tsx`
- **Routes**:
  - `/login` - Login page
  - `/discover` - Discover page
  - `/movie/:id` - Movie detail
  - `/tv/:id` - TV show detail

### Phase 4: Testing

#### 12. Playwright Test Suite
- **File**: `web/e2e/media-manager.spec.ts`
- **Tests**:
  - Auth: Login page visibility, toggle register
  - Navigation: All pages accessible
  - Discover: Page load, tab filtering, movie cards
  - Movie Detail: Navigation, info sidebar, similar movies
  - TV Detail: Navigation, season selector
  - Search: Search functionality
  - Responsive: Mobile and tablet viewports
  - Accessibility: Heading structure, clickable links

## 🎨 Design System

### Spotify-Inspired Dark Theme
- Background: `#121212` (near black)
- Surface: `#1f1f1f` (dark card)
- Accent: `#1ed760` (Spotify green)
- Text: `#ffffff` (white), `#b3b3b3` (silver)
- Border radius: Pill buttons (9999px), Cards (8px)
- Typography: DM Sans font family
- Shadows: Heavy on dark backgrounds

## 🗄️ Database Schema

### New Tables (Auto-migrated)
- `users` - User accounts
- `watch_histories` - Viewing progress
- `watchlists` - Saved content
- `user_preferences` - User settings
- `tv_shows` - TV series metadata
- `seasons` - TV seasons
- `episodes` - TV episodes
- `cast_members` - Actors
- `crew_members` - Directors/writers
- `genres` - Genre categories
- `videos` - Trailers/clips

## 🔌 API Endpoints

### Auth
- `POST /api/auth/register` - Create account
- `POST /api/auth/login` - Sign in
- `GET /api/auth/me` - Get current user (protected)
- `POST /api/auth/logout` - Sign out

### Discover
- All endpoints listed in Phase 1.4

## 📦 Dependencies Added

### Backend
- `github.com/golang-jwt/jwt/v5` - JWT authentication

### Frontend
- Already had: React, React Router, TanStack Query, Framer Motion, Tailwind
- Uses: Lucide React icons

## 🚀 Next Steps (Remaining Features)

### Pending Phase 2: Streaming
- [ ] Video streaming endpoint (HLS with FFmpeg)
- [ ] Video player component with controls
- [ ] Subtitle support

### Pending Phase 3: Enhanced Discovery
- [ ] Search filters (year, rating, genre)
- [ ] Watch history tracking
- [ ] Resume points
- [ ] Watchlist functionality

### Pending Phase 4: TV Optimization
- [ ] 10-foot UI for Nvidia Shield
- [ ] D-pad navigation support
- [ ] Android TV launcher integration
- [ ] Voice search

## 🧪 Running Tests

```bash
# Install Playwright browsers
npx playwright install

# Run tests
npx playwright test

# Run with UI
npx playwright test --ui

# Run specific test
npx playwright test media-manager.spec.ts
```

## 📝 Notes

- The JWT secret should be changed in production (`internal/service/auth.go`)
- TMDB API key is required for discover features
- The app uses the existing TMDb service integration
- All new models are auto-migrated on startup
- The frontend uses the existing React + Tailwind setup
