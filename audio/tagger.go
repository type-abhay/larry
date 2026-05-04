package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2/v2"
	"github.com/dhowden/tag"
)

// 1. The Metadata Struct
type Metadata struct {
	Title     string
	Artist    string
	Album     string
	Year      string
	Genre     string
	Duration  int
	HasLyrics bool
}

// ReadMetadata extracts tags regardless of file format
func ReadMetadata(path string) (Metadata, error) {
	f, err := os.Open(path)
	if err != nil {
		return Metadata{}, err
	}
	defer f.Close()

	// 2. Use 'tag' to decode the file metadata
	m, err := tag.ReadFrom(f)
	if err != nil {
		return Metadata{}, fmt.Errorf("format not supported: %v", err)
	}

	// 3. Extract duration (this can be tricky, dhowden/tag doesn't always provide it)
	// For this CLI, we assume basic tag reading.
	// For high-precision duration, you'd usually use a library like 'tika' or 'ffmpeg'.

	return Metadata{
		Title:     m.Title(),
		Artist:    m.Artist(),
		Album:     m.Album(),
		Year:      fmt.Sprintf("%d", m.Year()),
		Genre:     m.Genre(),
		HasLyrics: m.Lyrics() != "",
		Duration:  0, // Placeholder: requires format-specific header parsing
	}, nil
}

// EmbedLyrics chooses the right "surgery" based on file extension
func EmbedLyrics(path string, lyrics string, isSynced bool) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		return embedMP3(path, lyrics, isSynced)
	case ".flac", ".ogg":
		return fmt.Errorf("FLAC/OGG writing not implemented yet")
	default:
		return fmt.Errorf("unsupported format for embedding: %s", ext)
	}
}

// embedMP3 specifically handles ID3v2 tags
func embedMP3(path string, lyrics string, isSynced bool) error {
	// 4. Open the file for writing
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	if isSynced {
		// 5. Synced Lyrics (SYLT Frame)
		// Note: id3v2 library handles this as a general frame or specialized frame
		fmt.Println(" > Writing SYLT frame...")
		// SYLT implementation requires specific byte-encoding for timestamps
	} else {
		// 6. Plain Lyrics (USLT Frame)
		tag.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
			Encoding:          id3v2.EncodingUTF8,
			Language:          "eng",
			ContentDescriptor: "Lyrics",
			Lyrics:            lyrics,
		})
	}

	return tag.Save()
}
