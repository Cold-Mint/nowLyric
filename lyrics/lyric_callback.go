package lyrics

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

type LyricCallback struct {
	OnlyTranslation bool
	OutputPath      string
	lastLine        string
}

func (lc *LyricCallback) Play(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Stop(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Paused(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) UpdateLyric(playerBusName, line string, progress float64, lyric *Lyric) {
	str := line
	if lc.OnlyTranslation {
		if idx := strings.Index(str, "  "); idx > -1 {
			str = str[idx+2:]
		}
	}
	if lc.lastLine == str {
		return
	}
	lc.lastLine = str
	println(str)
	if lc.OutputPath != "" {
		_ = WriteToFile(lc.OutputPath, str) // 忽略错误，或打印日志
	}
}

// The directory is created only once 目录只创建一次
var dirOnce sync.Once
var dirErr error

func WriteToFile(outputPath, content string) error {
	dirOnce.Do(func() {
		dirErr = os.MkdirAll(filepath.Dir(outputPath), 0755)
	})
	if dirErr != nil {
		return dirErr
	}
	return os.WriteFile(outputPath, unsafe.Slice(unsafe.StringData(content), len(content)), 0644)
}
