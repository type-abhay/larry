package audio

import (
	"io/fs"
	"path/filepath"
	"strings"
)

// 1. Supported Extensions
// We define this as a map for O(1) "fast lookups"
var supportedExtensions = map[string]bool{
	".mp3":  true,
	".flac": true,
	".m4a":  true,
	".ogg":  true,
}

// ScanDirectory traverses the folder and subfolders to find music
// ScanDirectory traverses the folder and subfolders to find music
func ScanDirectory(root string) ([]string, error) {
	var audioFiles []string
	const maxFiles = 2000 // Your hard limit

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Check the limit
		if len(audioFiles) >= maxFiles {
			return filepath.SkipDir // Stop scanning once we hit 2000
		}

		ext := strings.ToLower(filepath.Ext(path))
		if supportedExtensions[ext] {
			audioFiles = append(audioFiles, path)
		}
		return nil
	})

	return audioFiles, err
}
