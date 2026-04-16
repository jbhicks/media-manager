# E2E Tests for Media Manager Web UI

This directory contains end-to-end tests for the Media Manager React web application using Playwright.

## Test Coverage

### Dashboard (`dashboard.spec.ts`)
- Dashboard title and description display
- Stats cards (Pending, Downloading, Completed, Failed)
- Quick action cards navigation (Downloads, Library, Search)
- Stats loading from API

### Downloads (`downloads.spec.ts`)
- Downloads page UI elements
- Clear Completed/Failed buttons
- Task list display with progress bars
- Task action buttons (Cancel, Restart, Delete)
- Task details (status, size, seeders, progress)
- Error message display

### Library (`library.spec.ts`)
- Library page UI elements
- Movie grid display with posters
- Movie metadata (title, size, rating)
- Clean Filenames functionality
- Empty state handling
- Image lazy loading

### Search (`search.spec.ts`)
- Search page UI elements
- Search input functionality
- Status filter dropdown
- Suggestion cards display
- Approve/Reject buttons
- Bulk selection and actions
- Generate Suggestions functionality

### Suggestions (`suggestions.spec.ts`)
- Suggestions page UI elements
- Stats cards (Pending, Approved, Rejected, Total)
- Generate Suggestions functionality
- Recent suggestions list
- Empty state handling

### Settings (`settings.spec.ts`)
- Settings page UI elements
- Download Sources card
- Download Rules card
- Server Settings card
- Security card
- VPN Status card

### Navigation & Theme (`navigation.spec.ts`)
- Header elements (title, menu, theme toggle, bell)
- Sidebar navigation to all pages
- Sidebar collapse/expand
- Theme toggle (dark/light mode)
- Active page highlighting

### VPN Status (`vpn.spec.ts`)
- VPN status display in header
- VPN connected/disconnected states
- VPN status card in Settings
- Auto-refresh functionality
- Mobile responsiveness

### Toast Notifications (`toast.spec.ts`)
- Toast display on actions
- Different toast types (success, error, info)
- Auto-dismiss after timeout
- Manual dismiss functionality
- Multiple toast stacking

## Running Tests

### Prerequisites
The backend server must be running on localhost:8080:
```bash
make dev
```

### Install Dependencies
```bash
cd web
npm install
npx playwright install
```

### Run All E2E Tests
```bash
npm run test:e2e
```

### Run Tests with UI Mode
```bash
npm run test:e2e:ui
```

### Run Tests in Debug Mode
```bash
npm run test:e2e:debug
```

### Run Tests for Specific Browser
```bash
npx playwright test --project=chromium
npx playwright test --project=firefox
npx playwright test --project=webkit
```

### Run Specific Test File
```bash
npx playwright test dashboard.spec.ts
```

## Configuration

Tests are configured in `playwright.config.ts`:
- Base URL: http://localhost:8080
- Browsers: Chromium, Firefox, WebKit
- Mobile devices: Pixel 5, iPhone 12
- Screenshots: On failure only
- Videos: On first retry

## Test Structure

Each test file follows this pattern:
1. Navigate to the page
2. Wait for API responses
3. Verify UI elements are visible
4. Test interactive features
5. Verify state changes

## VPN Status Feature

The VPN status feature:
- Checks VPN connection by querying the public IP
- Displays status in the header (green = connected, red = disconnected)
- Shows detailed VPN card in Settings with provider, location, and IP
- Auto-refreshes every 30 seconds
- Shows message explaining the status

### VPN Security Requirement

**Downloads are blocked when VPN is disconnected.** This is enforced at multiple levels:

1. **SearchAndDownload**: Checks VPN before starting any new downloads from search
2. **ProcessPendingDownloads**: Checks VPN before processing pending tasks
3. **startDownload**: Defense-in-depth check before adding torrent to client

When VPN is disconnected:
- A red warning banner appears on all pages
- All download operations are blocked
- Tasks remain in "pending" state until VPN reconnects
- Users see "VPN Disconnected" status in header and Settings

## API Endpoints Tested

- `/api/stats` - Dashboard statistics
- `/api/tasks` - Download tasks list
- `/api/library/movies` - Media library
- `/api/suggestions` - Download suggestions
- `/api/suggestions/stats` - Suggestion statistics
- `/api/suggestions/generate` - Generate suggestions
- `/api/sources` - Download sources
- `/api/rules` - Download rules
- `/api/library/reprocess` - Clean library filenames
- `/api/vpn/status` - VPN connection status
