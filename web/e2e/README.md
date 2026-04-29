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
- **Task actions (clicking Cancel, Restart, Delete)**
- Task details (status, size, seeders, progress, error messages)
- Empty state handling

### Library (`library.spec.ts`)
- Library page UI elements
- Movie grid display with posters
- Movie metadata (title, size, rating)
- Clean Filenames functionality
- Empty state handling
- Image lazy loading

### Search (`search.spec.ts`)
- Search page UI elements
- Search input with icon
- Search button functionality
- **Tab switching (Search Results / Suggestions)**
- **Search functionality with API integration**
- **Download buttons for search results**
- **Disabled download buttons when no magnet link**
- Search result metadata (size, seeders, leechers)
- Empty states for both tabs
- **Approve/Reject suggestions**
- **Image error handling with fallback**

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
- **VPN warning banner**
- **Dismiss VPN warning**

### VPN Actions (`vpn-actions.spec.ts`)
- **Connect VPN button**
- **Disconnect VPN button**
- **Connect/Disconnect API calls**
- **Loading states during VPN operations**
- **VPN status refresh after actions**
- **Button disabled states during operations**

### Toast Notifications (`toast.spec.ts`)
- Toast display on actions
- Different toast types (success, error, info)
- Auto-dismiss after timeout
- Manual dismiss functionality
- Multiple toast stacking

### Error States (`error-states.spec.ts`)
- **API failure handling**
- **404 error handling**
- **Network error handling**
- **Task action error handling**
- **Slow API response handling**
- **Error recovery**

### Accessibility (`accessibility.spec.ts`)
- **Heading structure**
- **Keyboard navigation**
- **Focus indicators**
- **ARIA labels**
- **Alt text on images**
- **Landmark regions**
- **Focus order**
- **Form labels**

### Responsive Design (`responsive.spec.ts`)
- **Mobile viewport (375x667)**
- **Tablet viewport (768x1024)**
- **Desktop viewport (1280x720)**
- **Grid layout adaptation**
- **Touch target sizes**
- **No horizontal scroll**
- **Landscape orientation**

### Theme (`theme.spec.ts`)
- **Default dark theme**
- **Theme toggle**
- **Theme persistence across navigation**
- **Theme persistence after reload**
- **Background and text colors**
- **Theme on all pages**

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

### Run Tests by Tag
```bash
npx playwright test --grep "accessibility"
```

## Configuration

Tests are configured in `playwright.config.ts`:
- Base URL: http://localhost:8080
- Browsers: Chromium, Firefox, WebKit
- Mobile devices: Pixel 5, iPhone 12
- Screenshots: On failure only
- Videos: On first retry
- Trace: On first retry

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
- `/api/tasks/cancel` - Cancel a task
- `/api/tasks/restart` - Restart a task
- `/api/tasks/delete` - Delete a task
- `/api/tasks/clear-completed` - Clear completed tasks
- `/api/tasks/clear-failed` - Clear failed tasks
- `/api/library/movies` - Media library
- `/api/library/reprocess` - Clean library filenames
- `/api/suggestions` - Download suggestions
- `/api/suggestions/stats` - Suggestion statistics
- `/api/suggestions/generate` - Generate suggestions
- `/api/suggestions/approve` - Approve suggestion
- `/api/suggestions/reject` - Reject suggestion
- `/api/search` - Search torrents
- `/api/sources` - Download sources
- `/api/rules` - Download rules
- `/api/vpn/status` - VPN connection status
- `/api/vpn/connect` - Connect VPN
- `/api/vpn/disconnect` - Disconnect VPN

## Best Practices

1. **Wait for API responses**: Always wait for API responses before asserting
2. **Conditional assertions**: Use conditional checks for data-dependent assertions
3. **Error handling**: Tests should handle empty states and errors gracefully
4. **Isolation**: Each test should be independent
5. **Cleanup**: No persistent state changes between tests