package lyrics

import (
	"fmt"
	"strings"
	"unsafe"
)

type LyricCallback struct {
	OnlyTranslation   bool
	WithLog           bool
	RichText          bool
	SupportExecute    bool
	lastLine          string
	PlayedTextColor   string
	UnplayedTextColor string
	Offset            float64
	MmapOK            bool
	Ptr               unsafe.Pointer
}

func (lc *LyricCallback) Play(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Stop(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) Paused(playerBusName string, audioFilePath string, lyric *Lyric) {
}

func (lc *LyricCallback) UpdateLyric(playerBusName, line string, progress float64, lyric *Lyric) {
	if lc.WithLog {
		fmt.Printf("LyricCallback{OnlyTranslation:%v, WithLog:%v, RichText:%v, SupportExecute:%v, lastLine:%q, PlayedTextColor:%q, UnplayedTextColor:%q, Offset:%f} progress=%f line=%q\n",
			lc.OnlyTranslation, lc.WithLog, lc.RichText, lc.SupportExecute,
			lc.lastLine, lc.PlayedTextColor, lc.UnplayedTextColor, lc.Offset,
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
		played := min(int(float64(total)*(progress+lc.Offset)), total)
		playedStr := string(runes[:played])
		unplayedStr := string(runes[played:])
		out = fmt.Sprintf(
			`<span foreground='%s'>%s</span>`+
				`<span foreground='%s'>%s</span>`,
			lc.PlayedTextColor, playedStr,
			lc.UnplayedTextColor, unplayedStr)
	} else {
		out = str
	}
	if lc.SupportExecute {
		out = "<executor.markup.true> " + out
	}
	if lc.lastLine == out {
		return
	}
	lc.lastLine = out
	println(out)
	if lc.MmapOK {
		WriteCString(lc.Ptr, out, Size)
	}
}
