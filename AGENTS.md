# Agent Guidelines for media-manager

This is a Go project based on the .gitignore configuration.

## Commands
- **Build**: `go build`

## Build Process
When `air` is running for auto-reloading, build messages are logged to `tmp/build.log`. Agents should examine this file for build output and errors.
- **Test**: `go test ./...`
- **Test single package**: `go test ./path/to/package`
- **Test with coverage**: `go test -cover ./...`
- **Lint**: `go vet ./...` or use `golangci-lint run` if available
- **Format**: `go fmt ./...`

## Code Style

### Important Note
Agents should never ask for permission to proceed with tasks. Always take action directly unless explicitly instructed otherwise.

### Service Auto-Reload
The service (`cmd/media-manager-service`) runs with `air` for auto-reloading during development. When you make changes to service code:
- `air` automatically detects changes and rebuilds the service binary
- The service restarts automatically
- Build output goes to `tmp/build.log`
- Service output goes to `tmp/service.log`
- **Wait 2-3 seconds after making changes** for the rebuild to complete

If you need to verify the service restarted, check:
```bash
tail -20 tmp/service.log  # Should show recent startup messages
```

Note: The GUI application (`cmd/media-manager`) is separate and does NOT use auto-reload.
- Follow standard Go conventions (gofmt, go vet)
- Use Go modules for dependency management
- Package names should be lowercase, single words
- Use camelCase for exported functions, PascalCase for types
- Error handling: always check and handle errors explicitly
- Use meaningful variable names, avoid abbreviations
- Keep functions small and focused on single responsibility
- Add comments for exported functions and types
- Use `context.Context` for cancellation and timeouts
- Prefer composition over inheritance

## Web UI Styling Guidelines

### CSS Framework Reference
The project uses **GitHub Primer CSS** as a reference for styling patterns. The Primer CSS repository is available as a git submodule in `reference/primer-css/`.

**Key Primer Resources**:
- **Source Files**: `reference/primer-css/src/` - Contains all component styles
- **Documentation**: `reference/primer-css/docs/` - Component documentation and examples
- **Variables**: `reference/primer-css/src/support/variables/` - Color schemes, spacing, typography

### Styling Principles
When working on the web UI (`web/index.html` or server-side HTML templates):

1. **Use Primer Patterns**: Reference Primer CSS components for:
   - Color schemes (dark mode friendly)
   - Spacing and layout patterns
   - Component design (buttons, cards, forms)
   - Typography hierarchy

2. **CSS Variables**: Define CSS custom properties (variables) for:
   - Colors (background, text, accent, borders)
   - Spacing (margins, padding, gaps)
   - Transitions and animations
   - Keep consistent with Primer's design system

3. **Component Structure**: Follow Primer's component patterns:
   - **Buttons**: `reference/primer-css/src/buttons/`
   - **Cards/Boxes**: `reference/primer-css/src/box/`
   - **Navigation**: `reference/primer-css/src/navigation/`
   - **Forms**: `reference/primer-css/src/forms/`
   - **Layout**: `reference/primer-css/src/layout/`

4. **Responsive Design**: 
   - Mobile-first approach
   - Use Primer's breakpoint patterns
   - Ensure touch-friendly targets (min 44px)

5. **Dark Mode**:
   - Use Primer's dark color schemes as reference
   - Ensure sufficient contrast ratios
   - Test all states (hover, active, disabled)

6. **Accessibility**:
   - Follow Primer's accessibility patterns
   - Proper ARIA labels and roles
   - Keyboard navigation support
   - Focus indicators

### Current Tech Stack
- **Frontend**: HTMX (no JavaScript framework)
- **Styling**: Custom CSS (no external CSS framework, Primer used as reference only)
- **Server**: Go HTTP handlers returning HTML partials

### HTMX-First Development Philosophy

**IMPORTANT**: This project uses **HTMX for all interactive features**. Minimize custom JavaScript whenever possible.

**When to use HTMX (preferred):**
- Navigation and routing (`hx-get`, `hx-push-url`)
- Form submissions (`hx-post`, `hx-put`, `hx-delete`)
- Dynamic content updates (`hx-target`, `hx-swap`)
- Polling and live updates (`hx-trigger="every 2s"`)
- Conditional loading (`hx-trigger="revealed"`)
- Event-driven updates (`hx-trigger="click, keyup"`)
- Server-sent events and WebSockets (`hx-sse`, `hx-ws`)

**When custom JavaScript is acceptable:**
- UI-only operations (animations, transitions)
- Client-side state management (active nav indicators)
- Toast notifications and alerts
- Browser APIs (localStorage, sessionStorage, clipboard)
- Third-party integrations (analytics, chat widgets)

**Anti-patterns to avoid:**
- ❌ Using `fetch()` or `XMLHttpRequest` when HTMX can handle it
- ❌ Manual DOM manipulation that HTMX can do declaratively
- ❌ Client-side routing when HTMX `hx-push-url` works
- ❌ Complex JavaScript state management
- ❌ Heavy JavaScript frameworks or libraries

**HTMX Routing Pattern (Server-Side Detection):**

The app uses smart routing that detects the `HX-Request` header:
- **HTMX requests** (with `HX-Request: true` header) → Returns HTML partial only
- **Direct browser requests** (no header) → Returns full page with layout

Example server-side handler:
```go
func (s *HTTPServer) handleSuggestionsPage(w http.ResponseWriter, r *http.Request) {
    if r.Header.Get("HX-Request") == "true" {
        // Return partial for HTMX to swap
        s.handleSuggestionsPartial(w, r)
        return
    }
    // Return full page with layout for direct access
    s.serveFullPage(w, r, "suggestions")
}
```

Example navigation links:
```html
<!-- GOOD: Clean URLs with HTMX -->
<a href="/suggestions" 
   hx-get="/suggestions" 
   hx-target="#content"
   hx-push-url="true">
   Suggestions
</a>

<!-- AVOID: Hash routing or manual JavaScript -->
<a href="#suggestions" onclick="navigate('suggestions')">Suggestions</a>
```

**Benefits of this approach:**
- ✅ Clean, shareable URLs (`/suggestions` not `#suggestions`)
- ✅ Page refresh works correctly (loads full page with styling)
- ✅ Browser back/forward buttons work automatically
- ✅ HTMX handles all routing declaratively
- ✅ No client-side routing JavaScript needed

**IMPORTANT: Use Absolute Paths for Static Assets**

When using HTMX routing with different URL paths, always use **absolute paths** for CSS, JavaScript, and images:

```html
<!-- GOOD: Absolute paths work from any route -->
<link rel="stylesheet" href="/web/styles.css">
<script src="/web/app.js"></script>
<img src="/web/images/logo.png">

<!-- BAD: Relative paths break on different routes -->
<link rel="stylesheet" href="styles.css">        <!-- Works on /web/index.html, fails on /suggestions -->
<script src="app.js"></script>                   <!-- Works on /web/index.html, fails on /suggestions -->
<img src="images/logo.png">                       <!-- Works on /web/index.html, fails on /suggestions -->
```

Why: When the page URL changes (e.g., from `/web/index.html` to `/suggestions`), relative paths resolve differently:
- At `/web/index.html`: `styles.css` → `/web/styles.css` ✅
- At `/suggestions`: `styles.css` → `/suggestions/styles.css` ❌

### Example Usage
When creating a new UI component:
```css
/* Look at reference/primer-css/src/buttons/button.scss for patterns */
.btn-primary {
    /* Use Primer's color and spacing patterns */
    background: var(--accent-blue);
    padding: 8px 16px;
    border-radius: 6px;
    /* ... */
}
```

## Fyne GUI Development

When making any changes to the GUI, you must follow these rules to avoid threading issues:

- **All UI updates must run on the main thread.**
- Use `fyne.CurrentApp().Driver().Do()` to schedule UI operations on the main thread.
- **Do not use `time.AfterFunc` for UI updates.** This will cause the app to crash.

Incorrect:

`time.AfterFunc(1*time.Millisecond, func() {`
	`// UI update code here`
`})`

Correct:

`fyne.CurrentApp().Driver().Do(func() {`
	`// UI update code here`
`})`

## Testing Guidelines
- Use Go's built-in `testing` package for unit tests.
- Leverage `fyne.io/fyne/v2/test` for testing graphical components.
- Create `_test.go` files for test definitions.
- Simulate user interactions using `test.Type` and validate GUI behavior.
- Ensure tests run without displaying windows or requiring a GUI.
- Run tests using `go test ./...`.

---

## Main Function and Media Path Handling

- The main entry point is in `cmd/media-manager/main.go` (see line 10).
- The app accepts a directory path as the first argument (`os.Args[1]`). If not provided, it defaults to the current working directory.
- The selected directory is logged with `log.Printf("Opening directory: %s", dir)`.
- The directory path is **not** currently passed to the app logic (`app.NewMediaManagerApp()`), so the rest of the app does not use it.

## Media Loading

- The UI (`internal/ui/views/main.go`) currently only displays placeholder images and does not load real media files.
- The config (`internal/config/config.go`) does not use the directory argument from main.

## Recommendation

- To support dynamic media directories, pass the `dir` argument from main into the app and config layers, and use it for media loading.
- Add debug logging in main and wherever the media directory is used.

---

## Consolidated Agent-Oriented Content

### Testing Plan
- Use Go's built-in testing package (`testing`) for unit and integration tests.
- Leverage the `fyne.io/fyne/v2/test` package for graphical application testing.
- Structure tests to cover:
  - Core functionality of the media manager.
  - Edge cases and error handling.
  - Performance benchmarks.
  - GUI interactions using Fyne's testing utilities.

### TODOs for Agents
- Optimize thumbnail generation for large media libraries.
- Implement advanced tagging features (bulk tagging, color-coded tags).
- Develop robust sorting and filtering options (by date, size, tags, etc.).
- Enhance UI/UX for media browsing.
- Add support for additional media formats.
- Improve thumbnail caching mechanism.

### Development Guide Highlights
- Initialize and run:
  ```bash
  go mod tidy
  go run cmd/media-manager/main.go
  ```
- Development commands:
  ```bash
  go test ./...                           # Run all tests
  go test ./internal/scanner              # Test specific package
  go build -o bin/media-manager cmd/media-manager/main.go
  ```
- Key technologies:
  - **Framework**: Fyne v2 (native desktop GUI)
  - **Database**: SQLite with GORM ORM
  - **Thumbnails**: Go image packages, FFmpeg for videos
  - **File Watching**: fsnotify for real-time updates
  - **Testing**: Go standard testing, testify for assertions
  - **Web UI**: HTMX for dynamic interactions, Primer CSS as styling reference

---

## Browser Testing and Performance

When analyzing website performance, debugging web applications, or automating browser interactions, use the `chrome-devtools` MCP tools.

**Available capabilities:**
- **Performance analysis**: Record traces and extract actionable performance insights
- **Browser automation**: Reliable automation with puppeteer (click, fill forms, navigate)
- **Network debugging**: Analyze network requests, check browser console
- **Visual debugging**: Take screenshots and DOM snapshots
- **Emulation**: Test different devices, viewports, and network conditions

**Usage pattern:**
```
use chrome-devtools to check the performance of https://example.com
use chrome-devtools to take a screenshot of https://example.com
use chrome-devtools to analyze network requests for https://example.com
```

**Note**: The Chrome DevTools MCP server will automatically start a Chrome instance when needed. Always reference it explicitly in prompts when browser automation or performance analysis is required.

---

_Discovered by opencode agent, 2025-06-29._