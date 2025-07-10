
### Fyne Threading and UI Updates

- **All Fyne UI updates must be performed on the main thread.**
- If you need to update the UI from a background goroutine (e.g., after processing, file I/O, or ffmpeg), you **must** wrap the update in:
  ```go
  fyne.Do(func() {
      // UI update code here
  })
  ```
- Never update the UI directly from a goroutine or use `time.AfterFunc` for UI changes. This will cause the app to crash.
- Use `fyne.DoAndWait` if you need to block until the UI update is complete.
- Example (correct):
  ```go
  go func() {
      // ... background work ...
      fyne.Do(func() {
          myWidget.Refresh()
      })
  }()
  ```
- Example (incorrect):
  ```go
  go func() {
      myWidget.Refresh() // This will crash!
  }()
  ```

