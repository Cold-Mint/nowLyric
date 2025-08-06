package lyrics

import (
	"os"
	"path/filepath"
	"strings"
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

func (lc *LyricCallback) UpdateLyric(playerBusName string, line string, progress float64, lyric *Lyric) {
	str := line
	if lc.OnlyTranslation {
		sep := "  "
		idx := strings.Index(str, sep)
		if idx > -1 {
			str = str[idx+len(sep):]
		}
	}
	if lc.lastLine != str {
		lc.lastLine = str
		println(str)
		if strings.TrimSpace(lc.OutputPath) != "" {
			err := WriteToFile(lc.OutputPath, str)
			if err != nil {
				return
			}
		}
	}
}

// WriteToFile 将 content 写入指定文件
// 如果目录不存在会自动创建
func WriteToFile(outputPath string, content string) error {
	// 确保目录存在
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件（覆盖式写入）
	err := os.WriteFile(outputPath, []byte(content), 0644)
	return err
}
