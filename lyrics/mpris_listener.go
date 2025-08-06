package lyrics

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	ffmpeggo "github.com/u2takey/ffmpeg-go"
)

// MPrisListener Media Player Remote Interfacing Specification Listener
// 媒体播放器远程接口监听器
type MPrisListener struct {
	conn          *dbus.Conn
	CallBack      MusicEventCallback
	playing       bool
	lyric         *Lyric
	playerBusName string
}

// ConnectSessionBus connects to the session bus.
// 连接到会话总线
func (watcher *MPrisListener) ConnectSessionBus(withLog bool) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		if withLog {
			log.Fatal("D-Bus connect error:", err)
		}
		return err
	}
	err = conn.AddMatchSignal(
		dbus.WithMatchInterface("org.freedesktop.DBus.Properties"),
		dbus.WithMatchMember("PropertiesChanged"),
	)
	if err != nil {
		if withLog {
			log.Fatal("addMatchSignal failed:", err)
		}
		return err
	}
	if withLog {
		fmt.Println("Listening for MPris metadata or status changes...")
	}
	watcher.conn = conn
	return nil
}

// SynchronizedLyrics Synchronized lyrics
// 同步歌词
func (watcher *MPrisListener) SynchronizedLyrics(withLog bool, delay uint32) {
	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
	defer ticker.Stop()
	for range ticker.C {
		if !watcher.playing {
			continue
		}
		pos, err := watcher.getPosition()
		if err != nil {
			if withLog {
				println("Failed to get playback position:", err.Error())
			}
			continue
		}
		line, progress := watcher.lyric.LineAt(pos)
		if withLog {
			println("[DEBUG] Current lyric line:", line, progress)
		}
		if watcher.CallBack != nil {
			watcher.CallBack.UpdateLyric(watcher.playerBusName, line, progress, watcher.lyric)
		}
	}
}

func (watcher *MPrisListener) WatchPlayerEvents(withLog bool) {
	defer func(conn *dbus.Conn) {
		err := conn.Close()
		if err != nil {
			if withLog {
				log.Println("close dbus error:", err)
			}
		}
	}(watcher.conn)

	ch := make(chan *dbus.Signal, 16)
	watcher.conn.Signal(ch)
	if withLog {
		log.Println("Signal channel created, start listening")
	}

	for sig := range ch {
		if !watcher.isMarisSignal(sig, withLog) {
			continue
		}
		watcher.handleSignal(sig, withLog)
	}
}

func (watcher *MPrisListener) isMarisSignal(sig *dbus.Signal, withLog bool) bool {
	if !strings.HasPrefix(string(sig.Path), "/org/mpris/MediaPlayer2") {
		if withLog {
			log.Println("[DEBUG] Signal path does not match MPRIS prefix, skipping")
		}
		return false
	}
	if len(sig.Body) < 2 {
		if withLog {
			log.Println("[DEBUG] Signal body length less than 2, skipping")
		}
		return false
	}
	return true
}

func (watcher *MPrisListener) handleSignal(sig *dbus.Signal, withLog bool) {
	props, ok := sig.Body[1].(map[string]dbus.Variant)
	if !ok {
		if withLog {
			log.Println("[WARN] Failed to cast signal.Body[1] to map[string]dbus.Variant")
		}
		return
	}

	if path := watcher.extractLocalPath(props, withLog); path != "" {
		watcher.onAudioFileChanged(path, sig.Sender, withLog)
	}

	if status := watcher.extractStatus(props); status != "" {
		watcher.onPlaybackStatusChanged(status, withLog)
	}
}

func (watcher *MPrisListener) extractLocalPath(props map[string]dbus.Variant, withLog bool) string {
	metaVar, ok := props["Metadata"]
	if !ok {
		return ""
	}
	meta := metaVar.Value().(map[string]dbus.Variant)
	urlVar, ok := meta["xesam:url"]
	if !ok {
		return ""
	}
	urlStr := urlVar.Value().(string)
	if withLog {
		log.Printf("[DEBUG] Metadata xesam:url = %s\n", urlStr)
	}
	if !strings.HasPrefix(urlStr, "file://") {
		return ""
	}
	local := strings.TrimPrefix(urlStr, "file://")
	decoded, err := url.PathUnescape(local)
	if err != nil {
		if withLog {
			log.Printf("[ERROR] Failed to decode file path: %v\n", err)
		}
		return ""
	}
	if withLog {
		log.Printf("[DEBUG] Decoded file path: %s\n", decoded)
	}
	if !isAudioFile(decoded) {
		if withLog {
			log.Printf("[DEBUG] File is not recognized audio file: %s\n", decoded)
		}
		return ""
	}
	return decoded
}

func (watcher *MPrisListener) onAudioFileChanged(path, sender string, withLog bool) {
	watcher.playerBusName = sender
	lrcPath := strings.TrimSuffix(path, filepath.Ext(path)) + ".lrc"
	if withLog {
		log.Printf("[DEBUG] Looking for lyric file: %s\n", lrcPath)
	}
	if _, err := os.Stat(lrcPath); err != nil {
		if withLog {
			log.Printf("[WARN] Lyric file not found: %s\n", lrcPath)
		}
		return
	}
	err := watcher.getAllProperties()
	if err != nil {
		if withLog {
			log.Printf("[ERROR] Failed get all properties: %v\n", err)
		}
		return
	}
	dur, err := watcher.getSongDuration(path)
	if err != nil {
		if withLog {
			log.Printf("[ERROR] Failed to get song duration: %v\n", err)
		}
		return
	}
	watcher.lyric, err = NewLyric(lrcPath, dur)
	if err != nil {
		if withLog {
			log.Printf("[ERROR] Failed to parse lyric file %s: %v\n", lrcPath, err)
		}
	} else if withLog {
		log.Printf("[INFO] Loaded lyric file: %s\n", lrcPath)
	}
}

func (watcher *MPrisListener) extractStatus(props map[string]dbus.Variant) string {
	sv, ok := props["PlaybackStatus"]
	if !ok {
		return ""
	}
	return sv.Value().(string)
}

func (watcher *MPrisListener) onPlaybackStatusChanged(status string, withLog bool) {
	if watcher.playerBusName == "" {
		return
	}
	switch status {
	case "Playing":
		watcher.playing = true
		if withLog {
			log.Printf("[INFO] Triggering Play callback for bus: %s\n", watcher.playerBusName)
		}
		if watcher.CallBack != nil {
			watcher.CallBack.Play(watcher.playerBusName, "", watcher.lyric)
		}
	case "Stopped":
		watcher.playing = false
		if withLog {
			log.Printf("[INFO] Triggering Stop callback for bus: %s\n", watcher.playerBusName)
		}
		if watcher.CallBack != nil {
			watcher.CallBack.Stop(watcher.playerBusName, "", watcher.lyric)
		}
	case "Paused":
		watcher.playing = false
		if withLog {
			log.Printf("[INFO] Triggering Paused callback for bus: %s\n", watcher.playerBusName)
		}
		if watcher.CallBack != nil {
			watcher.CallBack.Paused(watcher.playerBusName, "", watcher.lyric)
		}
	default:
		if withLog {
			log.Printf("[WARN] Unknown playback status: %s\n", status)
		}
	}
}

func (watcher *MPrisListener) getAllProperties() error {
	if watcher.playerBusName == "" {
		return fmt.Errorf("no player bus name set")
	}
	obj := watcher.conn.Object(watcher.playerBusName, "/org/mpris/MediaPlayer2")
	var properties map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.mpris.MediaPlayer2.Player").Store(&properties)
	if err != nil {
		return fmt.Errorf("failed to get all properties: %v", err)
	}
	return nil
}

// 获取音频文件时长（单位：微秒）
func (watcher *MPrisListener) getSongDuration(audioFilePath string) (uint64, error) {
	// 使用 ffmpeg-go 的 Probe 方法获取音频文件的元数据，返回 JSON 格式的数据
	output, err := ffmpeggo.Probe(audioFilePath)
	if err != nil {
		return 0, fmt.Errorf("failed to probe the audio file: %v", err)
	}

	// 解析 JSON 数据
	var metadata map[string]interface{}
	err = json.Unmarshal([]byte(output), &metadata)
	if err != nil {
		return 0, fmt.Errorf("failed to parse ffmpeg probe output: %v", err)
	}

	// 从 "format" 键中获取音频文件的时长信息
	format := metadata["format"].(map[string]interface{})
	durationStr := format["duration"].(string)

	// 将时长转换为 float64 类型的秒数
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert duration to float: %v", err)
	}

	// 将时长从秒转换为微秒
	durationInMicroseconds := uint64(duration * 1_000_000)
	return durationInMicroseconds, nil
}

// 获取当前音乐的部分位置（微妙us，错误）
func (watcher *MPrisListener) getPosition() (uint64, error) {
	obj := watcher.conn.Object(watcher.playerBusName, "/org/mpris/MediaPlayer2")
	var variant dbus.Variant
	err := obj.Call("org.freedesktop.DBus.Properties.Get", 0,
		"org.mpris.MediaPlayer2.Player", "Position").Store(&variant)
	if err != nil {
		return 0, err
	}
	if val, ok := variant.Value().(int64); ok {
		return uint64(val), nil
	}
	return 0, fmt.Errorf("unexpected type for Position")
}
