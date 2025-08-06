package lyrics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

type LyricCallback struct {
	OnlyTranslation   bool
	WithLog           bool
	RichText          bool
	SupportExecute    bool
	OutputPath        string
	lastLine          string
	PlayedTextColor   string
	UnplayedTextColor string
	Offset            float64
}

func (lc *LyricCallback) Play(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Stop(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Paused(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) UpdateLyric(playerBusName, line string, progress float64, lyric *Lyric) {
	if lc.WithLog {
		fmt.Printf("LyricCallback{OnlyTranslation:%v, WithLog:%v, RichText:%v, SupportExecute:%v, OutputPath:%q, lastLine:%q, PlayedTextColor:%q, UnplayedTextColor:%q, Offset:%f} progress=%f line=%q\n",
			lc.OnlyTranslation, lc.WithLog, lc.RichText, lc.SupportExecute,
			lc.OutputPath, lc.lastLine, lc.PlayedTextColor, lc.UnplayedTextColor, lc.Offset,
			progress, line)
	}
	str := line
	if lc.OnlyTranslation {
		if idx := strings.Index(str, "  "); idx > -1 {
			str = str[idx+2:]
		}
	}
	var out string
	if lc.RichText {
		runes := []rune(str)
		total := len(runes)
		played := int(float64(total) * (progress + lc.Offset))
		playedStr := string(runes[:played])
		unplayedStr := string(runes[played:])
		out = fmt.Sprintf(
			` <span foreground='%s'>%s</span>`+
				` <span foreground='%s'>%s</span>`,
			lc.PlayedTextColor, playedStr,
			lc.UnplayedTextColor, unplayedStr)
	} else {
		out = str
	}
	if lc.SupportExecute {
		out = "<executor.markup.true>" + out
	}
	if lc.lastLine == out {
		return
	}
	lc.lastLine = out
	println(out)
	if lc.OutputPath != "" {
		_ = WriteToFile(lc.OutputPath, out)
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
