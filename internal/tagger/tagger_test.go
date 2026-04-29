package tagger

import (
	"os"
	"testing"
	"time"
)

func TestExtractTagsFromFilename(t *testing.T) {
	tagger := NewFilenameTagger(nil) // No database needed for parsing tests

	testCases := []struct {
		filename string
		expected ParsedFilename
	}{
		{
			filename: "RearEndParade.25.11.24.Brandy.Salazar.Family.1080p.HEVC.x265.PRT.mp4",
			expected: ParsedFilename{
				Studio:    "Rear End Parade",
				Date:      &[]time.Time{time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Brandy Salazar"},
				Title:     "",
				Format:    ".mp4",
				Quality:   []string{"FAMILY", "1080P", "HEVC", "X265", "PRT"},
			},
		},
		{
			filename: "BigChestCreamPie.25.12.13.Kali.Roses.mp4",
			expected: ParsedFilename{
				Studio:    "Big Chest Cream Pie",
				Date:      &[]time.Time{time.Date(2025, 12, 13, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Kali Roses"},
				Title:     "",
				Format:    ".mp4",
				Quality:   []string{},
			},
		},
		{
			filename: "PremiumExtra.24.11.18.Thick.Butt.Daphne.And.Kelsey.Kane.mp4",
			expected: ParsedFilename{
				Studio:    "Premium Extra",
				Date:      &[]time.Time{time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Daphne", "Kelsey Kane"},
				Title:     "Thick Butt",
				Format:    ".mp4",
				Quality:   []string{},
			},
		},
		{
			filename: "PremiumExtra.25.12.19.Mz.Dani.And.Uptown.Jenny.Curvy.Fistfights.Get.Energized.mp4",
			expected: ParsedFilename{
				Studio:    "Premium Extra",
				Date:      &[]time.Time{time.Date(2025, 12, 19, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Mz Dani", "Uptown Jenny"},
				Title:     "Curvy Fistfights Get Energized",
				Format:    ".mp4",
				Quality:   []string{},
			},
		},
		{
			filename: "AdvancedFights.25.10.31.Luna.Legend.Vs.Cody.Carter.mp4",
			expected: ParsedFilename{
				Studio:    "Advanced Fights",
				Date:      &[]time.Time{time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Luna Legend", "Cody Carter"},
				Title:     "Vs",
				Format:    ".mp4",
				Quality:   []string{},
			},
		},
		{
			filename: "MyLifeInFlorida.25.11.19.Sofie.Reyes.Sofie.Takes.It.All.Family.1080p.HEVC.x265.PRT.mp4",
			expected: ParsedFilename{
				Studio:    "My Life In Florida",
				Date:      &[]time.Time{time.Date(2025, 11, 19, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Sofie Reyes"},
				Title:     "Sofie Takes It All",
				Format:    ".mp4",
				Quality:   []string{"FAMILY", "1080P", "HEVC", "X265", "PRT"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result, err := tagger.ExtractTagsFromFilename(tc.filename)
			if err != nil {
				t.Fatalf("Failed to parse filename: %v", err)
			}

			// Check studio
			if result.Studio != tc.expected.Studio {
				t.Errorf("Studio mismatch: got %q, expected %q", result.Studio, tc.expected.Studio)
			}

			// Check date
			if tc.expected.Date != nil {
				if result.Date == nil {
					t.Errorf("Expected date %v, got nil", tc.expected.Date)
				} else if !result.Date.Equal(*tc.expected.Date) {
					t.Errorf("Date mismatch: got %v, expected %v", result.Date, tc.expected.Date)
				}
			}

			// Check actresses
			if len(result.Actresses) != len(tc.expected.Actresses) {
				t.Errorf("Actress count mismatch: got %d, expected %d", len(result.Actresses), len(tc.expected.Actresses))
			} else {
				for i, expected := range tc.expected.Actresses {
					if i < len(result.Actresses) && result.Actresses[i] != expected {
						t.Errorf("Actress %d mismatch: got %q, expected %q", i, result.Actresses[i], expected)
					}
				}
			}

			// Check format
			if result.Format != tc.expected.Format {
				t.Errorf("Format mismatch: got %q, expected %q", result.Format, tc.expected.Format)
			}
		})
	}
}

func TestExtractDetailedMetadata(t *testing.T) {
	tagger := NewFilenameTagger(nil)

	// Test with one of the sample video files
	testFile := "../../media/big_buck_bunny_480p_stereo [avidemux 30sec].mp4"

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skipf("Test video file not found: %s", testFile)
	}

	metadata, err := tagger.ExtractDetailedMetadata(testFile)
	if err != nil {
		t.Fatalf("Failed to extract metadata: %v", err)
	}

	// Basic checks - these will vary depending on the actual file
	if metadata.Width <= 0 {
		t.Error("Expected positive width")
	}
	if metadata.Height <= 0 {
		t.Error("Expected positive height")
	}
	if metadata.Duration <= 0 {
		t.Error("Expected positive duration")
	}

	// Check that we got some basic metadata
	t.Logf("Video codec: %s", metadata.VideoCodec)
	t.Logf("Resolution: %dx%d", metadata.Width, metadata.Height)
	t.Logf("Duration: %d seconds", metadata.Duration)
	t.Logf("Container: %s", metadata.ContainerFormat)
	t.Logf("Audio codec: %s", metadata.AudioCodec)
	t.Logf("Channels: %d", metadata.Channels)
}

func TestResolutionCategories(t *testing.T) {
	tagger := NewFilenameTagger(nil)

	testCases := []struct {
		width, height int
		expected      string
	}{
		{3840, 2160, "4K"},
		{1920, 1080, "1080p"},
		{1280, 720, "720p"},
		{640, 480, "480p"},
		{320, 240, "240p"},
	}

	for _, tc := range testCases {
		result := tagger.getResolutionCategory(tc.width, tc.height)
		if result != tc.expected {
			t.Errorf("Resolution %dx%d: got %s, expected %s", tc.width, tc.height, result, tc.expected)
		}
	}
}

func TestBitrateCategories(t *testing.T) {
	tagger := NewFilenameTagger(nil)

	testCases := []struct {
		bitrate  int
		expected string
	}{
		{60000, "50Mbps+"},
		{30000, "25-50Mbps"},
		{15000, "10-25Mbps"},
		{7000, "5-10Mbps"},
		{3000, "2-5Mbps"},
		{1500, "1-2Mbps"},
		{750, "500Kbps-1Mbps"},
		{300, "<500Kbps"},
	}

	for _, tc := range testCases {
		result := tagger.getBitrateCategory(tc.bitrate)
		if result != tc.expected {
			t.Errorf("Bitrate %d kbps: got %s, expected %s", tc.bitrate, result, tc.expected)
		}
	}
}

func TestExtractFriendlyTitle(t *testing.T) {
	tagger := NewFilenameTagger(nil)

	testCases := []struct {
		parsed   ParsedFilename
		expected string
	}{
		{
			parsed: ParsedFilename{
				Studio: "Rear End Parade",
				Date:   &[]time.Time{time.Date(2025, 11, 24, 0, 0, 0, 0, time.UTC)}[0],
				Title:  "",
			},
			expected: "Rear End Parade - 2025-11-24",
		},
		{
			parsed: ParsedFilename{
				Studio:    "Premium Extra",
				Date:      &[]time.Time{time.Date(2024, 11, 18, 0, 0, 0, 0, time.UTC)}[0],
				Actresses: []string{"Daphne", "Kelsey Kane"},
				Title:     "Thick Butt",
			},
			expected: "Premium Extra - Thick Butt - 2024-11-18",
		},
		{
			parsed: ParsedFilename{
				Studio: "Big Chest Cream Pie",
				Date:   &[]time.Time{time.Date(2025, 12, 13, 0, 0, 0, 0, time.UTC)}[0],
				Title:  "",
			},
			expected: "Big Chest Cream Pie - 2025-12-13",
		},
		{
			parsed: ParsedFilename{
				Studio: "",
				Date:   nil,
				Title:  "Some Random Title",
			},
			expected: "Some Random Title",
		},
		{
			parsed: ParsedFilename{
				Studio: "",
				Date:   nil,
				Title:  "",
			},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tagger.ExtractFriendlyTitle(&tc.parsed)
			if result != tc.expected {
				t.Errorf("Friendly title mismatch: got %q, expected %q", result, tc.expected)
			}
		})
	}
}
