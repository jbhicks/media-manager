package tagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/user/media-manager/internal/db"
	"github.com/user/media-manager/internal/ffmpeg"
	"github.com/user/media-manager/pkg/models"
)

// FilenameTagger extracts tags from video filenames
type FilenameTagger struct {
	database *db.Database
}

// NewFilenameTagger creates a new filename tagger
func NewFilenameTagger(database *db.Database) *FilenameTagger {
	return &FilenameTagger{
		database: database,
	}
}

// ParsedFilename represents the parsed components of a filename
type ParsedFilename struct {
	Studio    string
	Date      *time.Time
	Actresses []string
	Title     string
	Format    string
	Quality   []string // 1080p, HEVC, x265, etc.
}

// DetailedMetadata represents comprehensive metadata extracted from media files
type DetailedMetadata struct {
	// Video properties
	VideoCodec   string
	VideoBitrate int    // kbps
	FrameRate    string // e.g., "30/1", "60/1"
	Width        int
	Height       int
	Duration     int // seconds
	ColorSpace   string
	PixelFormat  string
	Profile      string // e.g., "Main", "High"

	// Audio properties
	AudioCodec    string
	AudioBitrate  int // kbps
	SampleRate    int // Hz
	Channels      int
	ChannelLayout string // e.g., "stereo", "5.1"

	// Container properties
	ContainerFormat string // mp4, mkv, avi, etc.
	FileSize        int64  // bytes

	// Image properties (for images)
	CameraModel  string
	ISO          int
	Aperture     string
	ShutterSpeed string
	FocalLength  string
	HasGPS       bool
}

// ExtractDetailedMetadata extracts comprehensive metadata from media files
func (t *FilenameTagger) ExtractDetailedMetadata(filePath string) (*DetailedMetadata, error) {
	metadata := &DetailedMetadata{}

	// Get basic info first
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	metadata.FileSize = info.Size()

	// Check if it's an image or video
	ext := strings.ToLower(filepath.Ext(filePath))
	isVideo := t.isVideoFile(ext)
	isImage := t.isImageFile(ext)

	if !isVideo && !isImage {
		return metadata, nil // Not a media file we can analyze
	}

	// Use ffprobe to get detailed stream information
	cmd, err := ffmpeg.NewFFprobeCommand(
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create ffprobe command: %w", err)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe failed: %w", err)
	}

	// Parse the JSON output
	var probeData struct {
		Format  map[string]interface{}   `json:"format"`
		Streams []map[string]interface{} `json:"streams"`
	}

	if err := json.Unmarshal(output, &probeData); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Extract format information
	if format, ok := probeData.Format["format_name"]; ok {
		if formatStr, ok := format.(string); ok {
			metadata.ContainerFormat = strings.Split(formatStr, ",")[0] // Take first format
		}
	}

	if duration, ok := probeData.Format["duration"]; ok {
		if durStr, ok := duration.(string); ok {
			if durFloat, err := strconv.ParseFloat(durStr, 64); err == nil {
				metadata.Duration = int(durFloat)
			}
		}
	}

	// Extract stream information
	for _, stream := range probeData.Streams {
		streamType, ok := stream["codec_type"].(string)
		if !ok {
			continue
		}

		switch streamType {
		case "video":
			if metadata.VideoCodec == "" { // Take first video stream
				if codec, ok := stream["codec_name"].(string); ok {
					metadata.VideoCodec = codec
				}
				if width, ok := stream["width"].(float64); ok {
					metadata.Width = int(width)
				}
				if height, ok := stream["height"].(float64); ok {
					metadata.Height = int(height)
				}
				if bitrate, ok := stream["bit_rate"].(string); ok {
					if br, err := strconv.Atoi(bitrate); err == nil {
						metadata.VideoBitrate = br / 1000 // Convert to kbps
					}
				}
				if fps, ok := stream["r_frame_rate"].(string); ok {
					metadata.FrameRate = fps
				}
				if colorSpace, ok := stream["color_space"].(string); ok {
					metadata.ColorSpace = colorSpace
				}
				if pixFmt, ok := stream["pix_fmt"].(string); ok {
					metadata.PixelFormat = pixFmt
				}
				if profile, ok := stream["profile"].(string); ok {
					metadata.Profile = profile
				}
			}
		case "audio":
			if metadata.AudioCodec == "" { // Take first audio stream
				if codec, ok := stream["codec_name"].(string); ok {
					metadata.AudioCodec = codec
				}
				if bitrate, ok := stream["bit_rate"].(string); ok {
					if br, err := strconv.Atoi(bitrate); err == nil {
						metadata.AudioBitrate = br / 1000 // Convert to kbps
					}
				}
				if sampleRate, ok := stream["sample_rate"].(string); ok {
					if sr, err := strconv.Atoi(sampleRate); err == nil {
						metadata.SampleRate = sr
					}
				}
				if channels, ok := stream["channels"].(float64); ok {
					metadata.Channels = int(channels)
				}
				if chLayout, ok := stream["channel_layout"].(string); ok {
					metadata.ChannelLayout = chLayout
				}
			}
		}
	}

	// For images, try to extract EXIF data
	if isImage {
		metadata.extractEXIFData(filePath)
	}

	return metadata, nil
}

// Example filename patterns:
// RearEndParade.25.11.24.Brandy.Salazar.Family.1080p.HEVC.x265.PRT.mp4
// BigChestCreamPie.25.12.13.Kali.Roses.mp4
// BigChestsRoundButts.25.12.06.Tia.Maria.mp4
// PremiumExtra.24.05.03.Rear.Call.The.Best.Of.Large.Butts.Family.1080p.HEVC.x265.PRT.mkv
// PremiumExtra.24.11.18.Thick.Butt.Daphne.And.Kelsey.Kane.mp4
// PremiumExtra.25.06.04.Brandy.Salazar.Trimz.Episode.4.mp4
// PremiumExtra.25.07.01.Abigail.Morris.Buying.Restraints.mp4
// PremiumExtra.25.11.30.Abigail.Morris.Touching.The.Award.Wife.mp4
// PremiumExtra.25.12.09.Jess.Nova.Grimy.Club.Romp.mp4
// PremiumExtra.25.12.19.Mz.Dani.And.Uptown.Jenny.Curvy.Fistfights.Get.Energized.mp4
// PremiumExtra.25.12.26.Abigail.Morris.Thick.Nurse.Loves.Thick.Rooster.mp4
// BirthdaySurprise.25.12.19.Vivian.Taylor.mp4
// DigitalExtraOriginals.25.12.19.Tessa.Thomas.And.Gia.Ohmy.mp4
// AdvancedFights.25.10.31.Luna.Legend.Vs.Cody.Carter.mp4
// IntenseX.25.11.22.Lucy.Lotus.Family.1080p.HEVC.x265.PRT.mp4
// HotwifeAdventures.25.12.17.Richelle.Ryan.mp4
// LittleDancers.24.11.18.Octokuro.Hiding.Her.Bikini.mp4
// MyLifeInFlorida.25.11.19.Sofie.Reyes.Sofie.Takes.It.All.Family.1080p.HEVC.x265.PRT.mp4
// MySillyFamily.25.10.04.Cassie.Lenoir.The.Next.Step.mp4
// OhMyMami.24.11.21.Devil.Khloe.mp4
// PremiumMediaLibrary.25.10.24.Zoey.Foxx.Intense.41781.mp4
// SensualMexico.25.09.24.Devil.Khloe.mp4
// SensualMex.25.10.18.Devil.Khloe.Family.1080p.HEVC.x265.PRT.mp4
// YouthCurves.25.12.22.Lucy.Lotus.mp4

// ExtractTagsFromFilename parses a filename and extracts tags
func (t *FilenameTagger) ExtractTagsFromFilename(filename string) (*ParsedFilename, error) {
	// Remove file extension
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	// Split by dots
	parts := strings.Split(name, ".")

	if len(parts) < 3 {
		return nil, fmt.Errorf("filename too short to parse: %s", filename)
	}

	parsed := &ParsedFilename{
		Format: filepath.Ext(filename),
	}

	// Look for date pattern (YY.MM.DD)
	dateIndex := -1
	for i, part := range parts {
		if t.isDatePart(part) {
			if i+2 < len(parts) && t.isDatePart(parts[i+1]) && t.isDatePart(parts[i+2]) {
				// Found YY.MM.DD pattern
				dateStr := fmt.Sprintf("%s.%s.%s", part, parts[i+1], parts[i+2])
				if date, err := t.parseDate(dateStr); err == nil {
					parsed.Date = date
					dateIndex = i
					break
				}
			}
		}
	}

	if dateIndex == -1 {
		// No date found, assume studio is first part
		parsed.Studio = t.cleanStudioName(parts[0])
		parsed.Actresses = t.extractActresses(parts[1:], 0) // Start from second part (after studio)
		parsed.Title = t.extractTitle(parts[1:], parsed.Actresses)
	} else {
		// Date found, studio is everything before date
		studioParts := parts[:dateIndex]
		parsed.Studio = t.cleanStudioName(strings.Join(studioParts, "."))

		// Actresses and title come after date
		remainingParts := parts[dateIndex+3:]
		parsed.Actresses = t.extractActresses(remainingParts, 0) // 0 means start from beginning of remainingParts
		parsed.Title = t.extractTitle(remainingParts, parsed.Actresses)
	}

	// Extract quality info
	parsed.Quality = t.extractQuality(parts)

	fmt.Printf("[DEBUG] Parsed %s -> Studio: '%s', Actresses: %v, Title: '%s', Date: %v\n",
		filename, parsed.Studio, parsed.Actresses, parsed.Title, parsed.Date)

	return parsed, nil
}

// isDatePart checks if a string could be a date component (YY, MM, or DD)
func (t *FilenameTagger) isDatePart(part string) bool {
	if len(part) != 2 {
		return false
	}
	num, err := strconv.Atoi(part)
	return err == nil && num >= 0 && num <= 99
}

// parseDate parses a YY.MM.DD date string
func (t *FilenameTagger) parseDate(dateStr string) (*time.Time, error) {
	parts := strings.Split(dateStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid date format")
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, err
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, err
	}

	// Assume 20xx for years 00-99
	if year < 100 {
		year += 2000
	}

	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return &date, nil
}

// cleanStudioName cleans up studio names
func (t *FilenameTagger) cleanStudioName(studio string) string {
	// First check for specific PG-13 studio name mappings
	studioLower := strings.ToLower(strings.ReplaceAll(studio, ".", ""))
	studioReplacements := map[string]string{
		"rearendparade":           "Rear End Parade",
		"bigchestcreampie":        "Big Chest Cream Pie",
		"bigchestsroundbutts":     "Big Chests Round Butts",
		"premiumextra":            "Premium Extra",
		"birthday surprise":       "Birthday Surprise",
		"digital extra originals": "Digital Extra Originals",
		"advancedfights":          "Advanced Fights",
		"intensex":                "Intense X",
		"hotwife adventures":      "Hotwife Adventures",
		"little dancers":          "Little Dancers",
		"mylifeinflorida":         "My Life In Florida",
		"my silly family":         "My Silly Family",
		"oh my mami":              "Oh My Mami",
		"premium media library":   "Premium Media Library",
		"sensualmexico":           "Sensual Mexico",
		"sensualmex":              "Sensual Mex",
		"youth curves":            "Youth Curves",
	}

	if replacement, exists := studioReplacements[studioLower]; exists {
		return replacement
	}

	// Convert to title case and handle concatenated words
	studio = strings.Title(strings.ToLower(studio))

	// Handle concatenated words - split on capital letters
	re := regexp.MustCompile(`([a-z])([A-Z])`)
	studio = re.ReplaceAllString(studio, `${1} ${2}`)

	return strings.TrimSpace(studio)
}

// extractActresses extracts actress names from filename parts
func (t *FilenameTagger) extractActresses(parts []string, startIdx int) []string {
	// Simple approach: collect all potential name parts, then filter out obvious title words
	var allParts []string

	for i := startIdx; i < len(parts); i++ {
		part := parts[i]
		partLower := strings.ToLower(part)

		// Stop if we hit quality/format indicators
		if t.isQualityPart(partLower) {
			break
		}

		// Collect potential name/title parts
		if !t.isNumeric(part) && len(part) > 1 && len(part) < 20 {
			allParts = append(allParts, part)
		}
	}

	// Now try to extract actresses from the collected parts
	return t.parseActressesFromParts(allParts)
}

// parseActressesFromParts tries to identify actress names from a sequence of parts
func (t *FilenameTagger) parseActressesFromParts(parts []string) []string {
	var actresses []string
	var currentActress []string
	seenInCurrent := make(map[string]bool)

	for _, part := range parts {
		partLower := strings.ToLower(part)

		// Check for "and" separator
		if partLower == "and" {
			if len(currentActress) > 0 {
				actresses = append(actresses, t.formatActressName(currentActress))
				currentActress = nil
				seenInCurrent = make(map[string]bool) // Reset for next actress
			}
			continue
		}

		// If we've seen this name part in the current actress, it might indicate repetition (title start)
		if seenInCurrent[partLower] && len(currentActress) > 1 {
			// We've seen this name part before in the current actress, likely the title is starting
			break
		}

		// Skip obvious title words
		if t.isObviousTitleWord(partLower) {
			if len(currentActress) > 0 {
				actresses = append(actresses, t.formatActressName(currentActress))
				currentActress = nil
				seenInCurrent = make(map[string]bool)
			}
			continue
		}

		currentActress = append(currentActress, part)
		seenInCurrent[partLower] = true
	}

	// Add the last actress
	if len(currentActress) > 0 {
		actresses = append(actresses, t.formatActressName(currentActress))
	}

	return actresses
}

// isObviousTitleWord checks if a word is clearly part of a title, not a name
func (t *FilenameTagger) isObviousTitleWord(word string) bool {
	obviousTitleWords := []string{
		"thick", "butt", "big", "curvy", "hot", "sexy", "wild", "naughty", "dirty",
		"vs", "versus", "the", "a", "an", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by",
		"episode", "part", "scene", "touching", "buying", "restraints", "award", "wife",
		"grimy", "club", "romp", "fistfights", "catfights", "get", "energized", "nurse", "loves", "rooster",
		"birthday", "surprise", "digital", "extra", "originals", "advanced", "fights", "intense", "x",
		"hotwife", "adventures", "little", "dancers", "my", "life", "florida", "silly", "family", "oh", "mami",
		"premium", "media", "library", "sensual", "mexico", "mex", "youth", "curves",
		"takes", "it", "all", // common title words
	}

	for _, titleWord := range obviousTitleWords {
		if word == titleWord {
			return true
		}
	}
	return false
}

// isTitleStart checks if this word indicates the start of the title section
func (t *FilenameTagger) isTitleStart(word string) bool {
	// Only consider it a title start if it's clearly not part of a name
	// and appears in contexts where titles typically begin
	titleStarters := []string{
		"the", "a", "an", "episode", "part", "scene",
		"touching", "buying", "grimy", "birthday", "advanced", "intense",
		"sofie", // repeated name in title - context dependent
	}

	word = strings.ToLower(word)
	for _, starter := range titleStarters {
		if word == starter {
			return true
		}
	}
	return false
}

// formatActressName formats an actress name from parts
func (t *FilenameTagger) formatActressName(parts []string) string {
	if len(parts) == 0 {
		return ""
	}

	// Capitalize each part
	for i, part := range parts {
		parts[i] = strings.Title(part)
	}

	return strings.Join(parts, " ")
}

// extractTitle extracts the title from remaining parts
func (t *FilenameTagger) extractTitle(parts []string, actresses []string) string {
	var titleParts []string

	// Create a set of actress name parts for filtering
	actressParts := make(map[string]bool)
	for _, actress := range actresses {
		for _, part := range strings.Split(strings.ToLower(actress), " ") {
			actressParts[part] = true
		}
	}

	for _, part := range parts {
		part = strings.ToLower(part)

		// Skip quality parts, actress parts, and "and"
		if t.isQualityPart(part) || actressParts[part] || part == "and" {
			continue
		}

		// Convert dots to spaces and capitalize
		part = strings.ReplaceAll(part, ".", " ")
		part = strings.Title(part)
		titleParts = append(titleParts, part)
	}

	return strings.Join(titleParts, " ")
}

// extractQuality extracts quality/format information
func (t *FilenameTagger) extractQuality(parts []string) []string {
	var quality []string

	qualityPatterns := []string{
		"1080p", "720p", "480p", "4k", "2160p",
		"hevc", "h264", "x265", "x264",
		"family", "prt", "mp4", "mkv", "avi",
		"intense",
	}

	for _, part := range parts {
		part = strings.ToLower(part)
		for _, pattern := range qualityPatterns {
			if part == pattern {
				quality = append(quality, strings.ToUpper(part))
				break
			}
		}
	}

	return quality
}

// isQualityPart checks if a part is quality/format related
func (t *FilenameTagger) isQualityPart(part string) bool {
	qualityParts := []string{
		"1080p", "720p", "480p", "4k", "2160p",
		"hevc", "h264", "x265", "x264",
		"family", "prt", "mp4", "mkv", "avi",
		"intense",
	}

	part = strings.ToLower(part)
	for _, qp := range qualityParts {
		if part == qp {
			return true
		}
	}
	return false
}

// isNumeric checks if a string is numeric
func (t *FilenameTagger) isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// TagMediaFile applies extracted tags to a media file
func (t *FilenameTagger) TagMediaFile(mediaFile *models.MediaFile) error {
	parsed, err := t.ExtractTagsFromFilename(mediaFile.Filename)
	if err != nil {
		// Log the error but continue with minimal tagging
		fmt.Printf("[WARN] Failed to parse filename %s: %v - using fallback\n", mediaFile.Filename, err)
		// Create a minimal parsed result
		parsed = &ParsedFilename{
			Format: filepath.Ext(mediaFile.Filename),
		}
	}

	// Extract detailed metadata
	metadata, err := t.ExtractDetailedMetadata(mediaFile.Path)
	if err != nil {
		// Log the error but continue with filename-based tagging
		fmt.Printf("[WARN] Failed to extract detailed metadata for %s: %v\n", mediaFile.Path, err)
	}

	var tags []models.Tag

	// Add filename-based tags
	// Add studio tag
	if parsed.Studio != "" {
		tag, err := t.getOrCreateTag("studio:" + parsed.Studio)
		if err != nil {
			return fmt.Errorf("failed to create studio tag: %w", err)
		}
		tags = append(tags, *tag)
	}

	// Add date tag
	if parsed.Date != nil {
		dateTag := fmt.Sprintf("date:%s", parsed.Date.Format("2006-01-02"))
		tag, err := t.getOrCreateTag(dateTag)
		if err != nil {
			return fmt.Errorf("failed to create date tag: %w", err)
		}
		tags = append(tags, *tag)
	}

	// Add actress tags
	for _, actress := range parsed.Actresses {
		if actress != "" {
			tag, err := t.getOrCreateTag("actress:" + actress)
			if err != nil {
				return fmt.Errorf("failed to create actress tag: %w", err)
			}
			tags = append(tags, *tag)
		}
	}

	// Add quality tags
	for _, quality := range parsed.Quality {
		tag, err := t.getOrCreateTag("quality:" + quality)
		if err != nil {
			return fmt.Errorf("failed to create quality tag: %w", err)
		}
		tags = append(tags, *tag)
	}

	// Add metadata-based tags
	if metadata != nil {
		// Video codec
		if metadata.VideoCodec != "" {
			tag, err := t.getOrCreateTag("codec:" + strings.ToUpper(metadata.VideoCodec))
			if err != nil {
				return fmt.Errorf("failed to create video codec tag: %w", err)
			}
			tags = append(tags, *tag)
		}

		// Audio codec
		if metadata.AudioCodec != "" {
			tag, err := t.getOrCreateTag("codec:" + strings.ToUpper(metadata.AudioCodec))
			if err != nil {
				return fmt.Errorf("failed to create audio codec tag: %w", err)
			}
			tags = append(tags, *tag)
		}

		// Resolution
		if metadata.Width > 0 && metadata.Height > 0 {
			resolution := t.getResolutionCategory(metadata.Width, metadata.Height)
			if resolution != "" {
				tag, err := t.getOrCreateTag("resolution:" + resolution)
				if err != nil {
					return fmt.Errorf("failed to create resolution tag: %w", err)
				}
				tags = append(tags, *tag)
			}
		}

		// Video bitrate category
		if metadata.VideoBitrate > 0 {
			bitrate := t.getBitrateCategory(metadata.VideoBitrate)
			if bitrate != "" {
				tag, err := t.getOrCreateTag("bitrate:" + bitrate)
				if err != nil {
					return fmt.Errorf("failed to create bitrate tag: %w", err)
				}
				tags = append(tags, *tag)
			}
		}

		// Container format
		if metadata.ContainerFormat != "" {
			tag, err := t.getOrCreateTag("format:" + strings.ToUpper(metadata.ContainerFormat))
			if err != nil {
				return fmt.Errorf("failed to create format tag: %w", err)
			}
			tags = append(tags, *tag)
		}

		// Audio channels
		if metadata.Channels > 0 {
			channels := t.getChannelCategory(metadata.Channels)
			if channels != "" {
				tag, err := t.getOrCreateTag("audio:" + channels)
				if err != nil {
					return fmt.Errorf("failed to create audio channels tag: %w", err)
				}
				tags = append(tags, *tag)
			}
		}

		// HDR detection (basic)
		if metadata.ColorSpace != "" && (strings.Contains(strings.ToLower(metadata.ColorSpace), "bt2020") || strings.Contains(strings.ToLower(metadata.PixelFormat), "10")) {
			tag, err := t.getOrCreateTag("quality:HDR")
			if err != nil {
				return fmt.Errorf("failed to create HDR tag: %w", err)
			}
			tags = append(tags, *tag)
		}
	}

	// Associate tags with media file
	mediaFile.Tags = tags

	// Set friendly title for display
	mediaFile.FriendlyTitle = t.ExtractFriendlyTitle(parsed)
	fmt.Printf("[DEBUG] Tagged %s -> FriendlyTitle: '%s'\n", mediaFile.Filename, mediaFile.FriendlyTitle)

	return t.database.GetDB().Save(mediaFile).Error
}

// getOrCreateTag gets an existing tag or creates a new one
func (t *FilenameTagger) getOrCreateTag(name string) (*models.Tag, error) {
	var tag models.Tag
	err := t.database.GetDB().Where("name = ?", name).First(&tag).Error
	if err != nil {
		// Tag doesn't exist, create it
		tag = models.Tag{
			Name:  name,
			Color: t.getTagColor(name),
		}
		err = t.database.GetDB().Create(&tag).Error
		if err != nil {
			return nil, err
		}
	}
	return &tag, nil
}

// getTagColor assigns colors based on tag type
func (t *FilenameTagger) getTagColor(tagName string) string {
	if strings.HasPrefix(tagName, "studio:") {
		return "#FF6B6B" // Red
	}
	if strings.HasPrefix(tagName, "actress:") {
		return "#4ECDC4" // Teal
	}
	if strings.HasPrefix(tagName, "date:") {
		return "#45B7D1" // Blue
	}
	if strings.HasPrefix(tagName, "quality:") {
		return "#96CEB4" // Green
	}
	if strings.HasPrefix(tagName, "codec:") {
		return "#FDCB6E" // Orange
	}
	if strings.HasPrefix(tagName, "resolution:") {
		return "#6C5CE7" // Purple
	}
	if strings.HasPrefix(tagName, "bitrate:") {
		return "#A29BFE" // Light purple
	}
	if strings.HasPrefix(tagName, "format:") {
		return "#FD79A8" // Pink
	}
	return "#FFEAA7" // Yellow (default)
}

// extractEXIFData extracts EXIF metadata from image files
func (m *DetailedMetadata) extractEXIFData(filePath string) {
	// For now, we'll use a simple approach with exiftool or imagemagick
	// This is a placeholder - in a real implementation you'd use a Go EXIF library
	// or call external tools like exiftool
	// For this demo, we'll leave it minimal
}

// isVideoFile checks if the file extension indicates a video file
func (t *FilenameTagger) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".mkv", ".webm", ".m4v", ".3gp", ".ogv", ".flv", ".asf", ".dv", ".mp2", ".mpg", ".mpeg", ".rm", ".wmv", ".ts", ".vob", ".divx"}
	ext = strings.ToLower(ext)
	for _, vext := range videoExts {
		if ext == vext {
			return true
		}
	}
	return false
}

// isImageFile checks if the file extension indicates an image file
func (t *FilenameTagger) isImageFile(ext string) bool {
	imageExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".tiff", ".bmp", ".ico", ".svg"}
	ext = strings.ToLower(ext)
	for _, iext := range imageExts {
		if ext == iext {
			return true
		}
	}
	return false
}

// getResolutionCategory categorizes resolution based on height
func (t *FilenameTagger) getResolutionCategory(width, height int) string {
	// Use the smaller dimension for categorization
	minDim := width
	if height < width {
		minDim = height
	}

	switch {
	case minDim >= 3840:
		return "8K"
	case minDim >= 2160:
		return "4K"
	case minDim >= 1440:
		return "1440p"
	case minDim >= 1080:
		return "1080p"
	case minDim >= 720:
		return "720p"
	case minDim >= 480:
		return "480p"
	case minDim >= 360:
		return "360p"
	case minDim >= 240:
		return "240p"
	default:
		return "SD"
	}
}

// getBitrateCategory categorizes bitrate in kbps
func (t *FilenameTagger) getBitrateCategory(bitrateKbps int) string {
	switch {
	case bitrateKbps >= 50000:
		return "50Mbps+"
	case bitrateKbps >= 25000:
		return "25-50Mbps"
	case bitrateKbps >= 10000:
		return "10-25Mbps"
	case bitrateKbps >= 5000:
		return "5-10Mbps"
	case bitrateKbps >= 2000:
		return "2-5Mbps"
	case bitrateKbps >= 1000:
		return "1-2Mbps"
	case bitrateKbps >= 500:
		return "500Kbps-1Mbps"
	default:
		return "<500Kbps"
	}
}

// ExtractFriendlyTitle creates a user-friendly display title from parsed filename
func (t *FilenameTagger) ExtractFriendlyTitle(parsed *ParsedFilename) string {
	if parsed == nil {
		return ""
	}

	// Build title in format: Studio - Description - Date
	var parts []string

	// Add studio if available
	if parsed.Studio != "" {
		parts = append(parts, parsed.Studio)
	}

	// Add main title/description
	if parsed.Title != "" {
		parts = append(parts, parsed.Title)
	}

	// Add date if available
	if parsed.Date != nil {
		dateStr := parsed.Date.Format("2006-01-02")
		parts = append(parts, dateStr)
	}

	// Join with " - " separator
	if len(parts) > 0 {
		return strings.Join(parts, " - ")
	}

	// Fallback: if no structured title, return the original title
	return parsed.Title
}

// getChannelCategory categorizes audio channels
func (t *FilenameTagger) getChannelCategory(channels int) string {
	switch channels {
	case 1:
		return "Mono"
	case 2:
		return "Stereo"
	case 6:
		return "5.1"
	case 8:
		return "7.1"
	default:
		return fmt.Sprintf("%dch", channels)
	}
}
