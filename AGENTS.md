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

---

_Discovered by opencode agent, 2025-06-29._