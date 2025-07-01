# Video Preview Plan

## Objective
Implement video previews for media files that play when the user hovers over a thumbnail in the UI.

## Approach
1. **Hover Detection**:
   - Use Fyne's event handling to detect hover events on media thumbnails.

2. **Video Preview Generation**:
   - Generate short video clips (e.g., 5 seconds) from the original media file using `ffmpeg`.
   - Store these clips in a temporary directory for playback.

3. **Playback Integration**:
   - Use an external library for video playback since Fyne does not natively support video.
   - Ensure seamless integration with Fyne's UI components.

4. **Fallback Mechanism**:
   - If video playback is not supported, display a static thumbnail instead.

5. **Performance Optimization**:
   - Pre-generate previews for visible thumbnails to reduce latency.
   - Cache previews to avoid redundant generation.

## Implementation Steps
1. **Hover Detection**:
   - Add hover event listeners to thumbnail components.

2. **Preview Generation**:
   - Extend the existing thumbnail generator to create short video clips.

3. **UI Integration**:
   - Embed video playback in the thumbnail area using an external library.

4. **Testing**:
   - Verify hover detection and video playback functionality.
   - Test performance and fallback behavior.

## Tools and Libraries
- **ffmpeg**: For video clip generation.
- **External Video Library**: TBD based on compatibility and performance.

## Challenges
- Ensuring smooth playback without UI lag.
- Handling unsupported formats gracefully.

## Timeline
1. Document plan: 1 day
2. Implement hover detection: 2 days
3. Generate video previews: 2 days
4. Integrate playback: 3 days
5. Testing and optimization: 2 days