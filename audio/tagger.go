package audio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	// Universal Reader
	"github.com/dhowden/tag"

	// MP3 Libraries
	"github.com/bogem/id3v2/v2"
	"github.com/tcolgate/mp3"

	// FLAC Libraries
	"github.com/go-flac/flacvorbis/v2"
	goflac "github.com/go-flac/go-flac/v2"

	// M4A Libraries
	"github.com/Sorrow446/go-mp4tag"
	"github.com/alfg/mp4"
)

// Metadata standardizes the tags across different file formats
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

	m, err := tag.ReadFrom(f)
	if err != nil {
		return Metadata{}, fmt.Errorf("format not supported for reading: %v", err)
	}

	duration := getDuration(path)

	return Metadata{
		Title:     m.Title(),
		Artist:    m.Artist(),
		Album:     m.Album(),
		Year:      fmt.Sprintf("%d", m.Year()),
		Genre:     m.Genre(),
		HasLyrics: m.Lyrics() != "",
		Duration:  duration,
	}, nil
}

// ==========================================
// DURATION ROUTER & LOGIC
// ==========================================

func getDuration(path string) int {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		return getMP3Duration(path)
	case ".flac":
		return getFLACDuration(path)
	case ".m4a":
		return getM4ADuration(path)
	default:
		return 0
	}
}

func getMP3Duration(path string) int {
	t, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer t.Close()

	d := mp3.NewDecoder(t)
	var totalDuration float64
	for {
		var f mp3.Frame
		if err := d.Decode(&f, nil); err != nil {
			if err == io.EOF {
				break
			}
			break
		}
		totalDuration += f.Duration().Seconds()
	}
	return int(totalDuration)
}

func getFLACDuration(path string) int {
	// Fixed: Using go-flac instead of mewkiz/flac to avoid API conflicts
	f, err := goflac.ParseFile(path)
	if err != nil {
		return 0
	}

	info, err := f.GetStreamInfo()
	if err != nil || info.SampleRate == 0 {
		return 0
	}

	return int(info.TotalSamples / uint64(info.SampleRate))
}

func getM4ADuration(path string) int {
	// Fixed: Manually handling the file descriptor so we can close it properly
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return 0
	}

	mp4File, err := mp4.OpenFromReader(f, info.Size())
	if err != nil {
		return 0
	}

	if mp4File.Moov != nil && mp4File.Moov.Mvhd != nil && mp4File.Moov.Mvhd.Timescale > 0 {
		// Fixed: Explicit type casting to uint64 before division
		return int(uint64(mp4File.Moov.Mvhd.Duration) / uint64(mp4File.Moov.Mvhd.Timescale))
	}
	return 0
}

// ==========================================
// EMBEDDING ROUTER & LOGIC
// ==========================================

func EmbedLyrics(path string, lyrics string, isSynced bool) error {
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".mp3":
		return embedMP3(path, lyrics, isSynced)
	case ".flac":
		return embedFLAC(path, lyrics, isSynced)
	case ".m4a":
		return embedM4A(path, lyrics, isSynced)
	default:
		return fmt.Errorf("unsupported format for embedding: %s", ext)
	}
}

func embedMP3(path string, lyrics string, isSynced bool) error {
	tag, err := id3v2.Open(path, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	tag.AddUnsynchronisedLyricsFrame(id3v2.UnsynchronisedLyricsFrame{
		Encoding:          id3v2.EncodingUTF8,
		Language:          "eng",
		ContentDescriptor: "Lyrics",
		Lyrics:            lyrics,
	})

	return tag.Save()
}

func embedFLAC(path string, lyrics string, isSynced bool) error {
	f, err := goflac.ParseFile(path)
	if err != nil {
		return err
	}

	var vorbis *flacvorbis.MetaDataBlockVorbisComment
	var vorbisIdx = -1

	for i, meta := range f.Meta {
		if meta.Type == goflac.VorbisComment {
			// Fixed: Dereferencing pointer using *meta
			vorbis, err = flacvorbis.ParseFromMetaDataBlock(*meta)
			if err != nil {
				return err
			}
			vorbisIdx = i
			break
		}
	}

	if vorbis == nil {
		vorbis = flacvorbis.New()
	}

	tagKey := "UNSYNCEDLYRICS"
	if isSynced {
		tagKey = "LYRICS"
	}

	vorbis.Add(tagKey, lyrics)
	newMeta := vorbis.Marshal()

	if vorbisIdx != -1 {
		// Fixed: Referencing as a pointer using &newMeta
		f.Meta[vorbisIdx] = &newMeta
	} else {
		// Fixed: Referencing as a pointer using &newMeta
		f.Meta = append(f.Meta, &newMeta)
	}

	return f.Save(path)
}

func embedM4A(path string, lyrics string, isSynced bool) error {
	mp4f, err := mp4tag.Open(path)
	if err != nil {
		return err
	}
	defer mp4f.Close()

	// Fixed: Extract the existing tags instead of calling SetLyrics directly
	tags, err := mp4f.Read()
	if err != nil {
		tags = &mp4tag.MP4Tags{}
	}

	// Fixed: Using the Custom map to manually insert the lyrics atom (©lyr)
	if tags.Custom == nil {
		tags.Custom = make(map[string]string)
	}
	tags.Custom["©lyr"] = lyrics

	// Fixed: Use Write instead of Save
	return mp4f.Write(tags, []string{})
}
