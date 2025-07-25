# Video Preview Generation Plan (Scene Overview)

- [x] Step 1: Locate the code responsible for generating animated previews (in `internal/preview/generator.go`).
- [x] Step 2: Identify the exact ffmpeg command being constructed and run.
- [x] Step 3: Check for problematic parameters (`-ss`, `-t`, `-frames`, filters).
- [x] Step 4: Analyze and explain why the current command fails for this file.
- [x] Step 5: Propose a corrected ffmpeg command that will work for this input.
- [x] Step 6: Redesign the preview generation to create a "scene summary" GIF:
    - [x] a. Decide on a strategy to sample frames evenly across the video duration (e.g., 10–20 short segments from throughout the video).
    - [x] b. For each segment, extract a short snippet (e.g., 0.5–1s) at regular intervals (e.g., every N% of the video).
    - [x] c. Concatenate these snippets into a single GIF, so the preview covers the whole video.
    - [x] d. Construct the ffmpeg command to extract and concatenate these segments efficiently (using `select`, `concat`, or `trim` filters, or by generating temporary files and combining them).
    - [x] e. Update the code to implement this new approach.
    - [x] f. Ensure the output GIF is not too large (limit total duration and resolution).
    - [x] g. Add comments and debug logging to clarify the new logic.
- [x] Step 7: Suggest code changes to implement the new "scene summary" preview logic.
