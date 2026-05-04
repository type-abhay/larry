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

	// Notice the third parameter: 'err error'
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		// 1. If there was an error accessing the path, handle it first
		if err != nil {
			return err
		}

		// 2. Skip directories
		if d.IsDir() {
			return nil
		}

		// 3. Check the file extension
		ext := strings.ToLower(filepath.Ext(path))
		if supportedExtensions[ext] {
			audioFiles = append(audioFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return audioFiles, nil
}
