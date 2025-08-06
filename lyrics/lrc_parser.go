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

type Lyric struct {
	Lines []LyricLine
	Maxus uint64 //The maximum length of the song, subtle
}
type LyricLine struct {
	TimeUs uint64 //Microsecond, a 64-bit unsigned integer
	Text   string //Lyrics text
}

// Song path, maximum length of the song
func NewLyric(path string, maxus uint64) (*Lyric, error) {
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
	return &Lyric{Lines: lines, Maxus: maxus}, nil
}

// 根据微妙数获取当播放的歌词
func (l *Lyric) GetLineUs(positionUs uint64) (string, float64) {
	if len(l.Lines) == 0 {
		return "", 0
	}

	// 查找当前歌词所在位置的索引
	idx := sort.Search(len(l.Lines), func(i int) bool {
		return l.Lines[i].TimeUs > positionUs
	})

	// 如果没有找到，则说明当前没有歌词
	if idx == 0 {
		return "", 0
	}

	// 获取当前歌词的起始时间
	currentLyric := l.Lines[idx-1]
	var nextLyric LyricLine

	// 如果已经是最后一句歌词，使用歌曲的最大长度作为结束时间
	if idx < len(l.Lines) {
		nextLyric = l.Lines[idx]
	} else {
		// 歌曲的结束时间就是最大时长
		nextLyric = LyricLine{TimeUs: l.Maxus}
	}

	// 计算当前歌词的播放进度
	timeDiff := nextLyric.TimeUs - currentLyric.TimeUs
	if timeDiff == 0 {
		return currentLyric.Text, 1 // 如果没有时间差，表示这句已经播放完
	}

	progress := float64(positionUs-currentLyric.TimeUs) / float64(timeDiff)
	//log.Printf("[ERROR] 当前时刻%d 结束时刻%d 最大长度%d\n", currentLyric.TimeUs, nextLyric.TimeUs, l.Maxus)

	// 返回当前歌词和进度
	return currentLyric.Text, progress
}
