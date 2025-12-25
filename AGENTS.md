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
Never run `make dev`. The user has it running in a separate process.
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

_Discovered by opencode agent, 2025-06-29._