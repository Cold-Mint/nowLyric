package lyrics

import (
	"bufio"
	_ "fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Lyric
// 歌词对象
type Lyric struct {
	Lines    []LyricLine
	Duration uint64 //The total duration of the song, with subtle units. 歌曲总时长，单位微妙。
	lastIdx  int
}

// LyricLine
// 歌词行对象
type LyricLine struct {
	TimeUs uint64 //Microsecond, a 64-bit unsigned integer
	Text   string //Lyrics text
}

// NewLyric Create the lyrics file object based on the file path and the duration of the audio file.
// 通过文件路径和音频文件时长来创建歌词文件对象。
func NewLyric(path string, duration uint64) (*Lyric, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	var lines []LyricLine
	scanner := bufio.NewScanner(file)
	timeTagRegex := regexp.MustCompile(`\[(\d+):(\d+\.\d+)]`)
	for scanner.Scan() {
		line := scanner.Text()
		tags := timeTagRegex.FindAllStringSubmatch(line, -1)
		text := timeTagRegex.ReplaceAllString(line, "")
		for _, tag := range tags {
			parseUint, _ := strconv.ParseUint(tag[1], 10, 64)
			secFloat, _ := strconv.ParseFloat(tag[2], 64)
			us := parseUint*60*1_000_000 + uint64(secFloat*1_000_000)
			lines = append(lines, LyricLine{TimeUs: us, Text: strings.TrimSpace(text)})
		}
	}
	sort.Slice(lines, func(i, j int) bool {
		return lines[i].TimeUs < lines[j].TimeUs
	})
	return &Lyric{Lines: lines, Duration: duration}, nil
}

// LineAt  Obtain the corresponding line lyrics based on the microseconds currently being played. How much has progress sung for the content of this line?
// 根据当前播放的微妙数获取对应的行歌词。progress为本行内容演唱了多少。
func (l *Lyric) LineAt(posUs uint64) (text string, progress float64) {
	if len(l.Lines) == 0 {
		return "", 0
	}
	if l.lastIdx >= 0 && l.lastIdx < len(l.Lines) {
		cur := l.Lines[l.lastIdx]
		nextUs := l.Duration
		if l.lastIdx+1 < len(l.Lines) {
			nextUs = l.Lines[l.lastIdx+1].TimeUs
		}
		if posUs >= cur.TimeUs && posUs < nextUs {
			return cur.Text, float64(posUs-cur.TimeUs) / float64(nextUs-cur.TimeUs)
		}
	}
	idx := sort.Search(len(l.Lines), func(i int) bool {
		return l.Lines[i].TimeUs > posUs
	})
	if idx == 0 {
		return "", 0
	}
	l.lastIdx = idx - 1
	cur := l.Lines[l.lastIdx]
	next := LyricLine{TimeUs: l.Duration}
	if idx < len(l.Lines) {
		next = l.Lines[idx]
	}
	return cur.Text, float64(posUs-cur.TimeUs) / float64(next.TimeUs-cur.TimeUs)
}
