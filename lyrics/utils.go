package lyrics

import (
	"path/filepath"
	"strings"
)

func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".mp3", ".flac", ".wav", ".m4a", ".aac", ".ogg", ".opus":
		return true
	default:
		return false
	}
}
