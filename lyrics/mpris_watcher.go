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

// 系统音频流的观察器
type MprisWatcher struct {
	conn *dbus.Conn
	//接口本身就是引用类型所以不用指针
	CallBack      MprisCallBack
	playing       bool
	lyric         *Lyric
	playerBusName string
}

// ConnectSessionBus connects to the session bus.
// 连接到会话总线
func (mprisWatcher *MprisWatcher) ConnectSessionBus(withLog bool) error {
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
		fmt.Println("Listening for MPRIS metadata or status changes...")
	}
	mprisWatcher.conn = conn
	return nil
}

func (mprisWatcher *MprisWatcher) SynchronizedLyrics(withLog bool, delay uint32) {
	duration := time.Duration(delay) * time.Millisecond
	for {
		time.Sleep(duration)
		if !mprisWatcher.playing {
			continue
		}
		pos, err := mprisWatcher.getPosition()
		if err != nil {
			if withLog {
				fmt.Println("Failed to get playback position:", err)
			}
			continue
		}
		line, progress := mprisWatcher.lyric.GetLineUs(pos)
		if withLog {
			fmt.Printf("[DEBUG] Current lyric line: %q(%f)\n", line, progress)
		}
		if mprisWatcher.CallBack != nil {
			mprisWatcher.CallBack.UpdateLyric(mprisWatcher.playerBusName, line, progress, mprisWatcher.lyric)
		}
	}
}

func (mprisWatcher *MprisWatcher) MonitorAudioChanges(withLog bool) {
	defer mprisWatcher.conn.Close()
	channel := make(chan *dbus.Signal, 16)
	mprisWatcher.conn.Signal(channel)
	var audioFilePath string
	if withLog {
		log.Println("Signal channel created, start listening")
	}
	for signal := range channel {
		if !strings.HasPrefix(string(signal.Path), "/org/mpris/MediaPlayer2") {
			if withLog {
				log.Println("[DEBUG] Signal path does not match MPRIS prefix, skipping")
			}
			continue
		}
		if len(signal.Body) < 2 {
			if withLog {
				log.Println("[DEBUG] Signal body length less than 2, skipping")
			}
			continue
		}

		changedProps, ok := signal.Body[1].(map[string]dbus.Variant)
		if !ok {
			if withLog {
				log.Println("[WARN] Failed to cast signal.Body[1] to map[string]dbus.Variant")
			}
			continue
		}

		if metadataVar, exists := changedProps["Metadata"]; exists {
			metadata := metadataVar.Value().(map[string]dbus.Variant)
			if urlVar, ok := metadata["xesam:url"]; ok {
				urlStr := urlVar.Value().(string)
				if withLog {
					log.Printf("[DEBUG] Metadata xesam:url = %s\n", urlStr)
				}
				if strings.HasPrefix(urlStr, "file://") {
					localPath := strings.TrimPrefix(urlStr, "file://")
					decodedPath, err := url.PathUnescape(localPath)
					if err != nil {
						if withLog {
							log.Printf("[ERROR] Failed to decode file path: %v\n", err)
						}
						continue
					}
					if withLog {
						log.Printf("[DEBUG] Decoded file path: %s\n", decodedPath)
					}
					if isAudioFile(decodedPath) {
						audioFilePath = decodedPath
						mprisWatcher.playerBusName = signal.Sender
						lrcPath := strings.TrimSuffix(decodedPath, filepath.Ext(decodedPath)) + ".lrc"
						if withLog {
							log.Printf("[DEBUG] Looking for lyric file: %s\n", lrcPath)
						}
						if _, err := os.Stat(lrcPath); err == nil {
							mprisWatcher.getAllProperties()
							var maxus, errorDuration = mprisWatcher.getSongDuration(decodedPath)
							if errorDuration != nil {
								if withLog {
									log.Printf("[ERROR] Failed to get song duration: %v\n", err)
								}
							}
							mprisWatcher.lyric, err = NewLyric(lrcPath, maxus)
							if err != nil {
								if withLog {
									log.Printf("[ERROR] Failed to parse lyric file %s: %v\n", lrcPath, err)
								}
							} else {
								if withLog {
									log.Printf("[INFO] Loaded lyric file: %s\n", lrcPath)
								}
							}
						} else {
							if withLog {
								log.Printf("[WARN] Lyric file not found: %s\n", lrcPath)
							}
						}
					} else {
						if withLog {
							log.Printf("[DEBUG] File is not recognized audio file: %s\n", decodedPath)
						}
					}
				}
			}
		}

		if statusVar, exists := changedProps["PlaybackStatus"]; exists {
			sender := signal.Sender
			if mprisWatcher.playerBusName != sender {
				if withLog {
					log.Printf("[DEBUG] Signal sender %s is not the current player %s, skipping\n", sender, mprisWatcher.playerBusName)
				}
				continue
			}
			status := statusVar.Value().(string)
			if withLog {
				log.Printf("[INFO] PlaybackStatus changed: %s\n", status)
			}
			switch status {
			case "Playing":
				mprisWatcher.playing = true
				if withLog {
					log.Printf("[INFO] Triggering Play callback for bus: %s, file: %s\n", mprisWatcher.playerBusName, audioFilePath)
				}
				if mprisWatcher.CallBack != nil {
					mprisWatcher.CallBack.Play(mprisWatcher.playerBusName, audioFilePath, mprisWatcher.lyric)
				}

			case "Stopped":
				mprisWatcher.playing = false
				if withLog {
					log.Printf("[INFO] Triggering Stop callback for bus: %s, file: %s\n", mprisWatcher.playerBusName, audioFilePath)
				}
				if mprisWatcher.CallBack != nil {
					mprisWatcher.CallBack.Stop(mprisWatcher.playerBusName, audioFilePath, mprisWatcher.lyric)
				}
			case "Paused":
				mprisWatcher.playing = false
				if withLog {
					log.Printf("[INFO] Triggering Paused callback for bus: %s, file: %s\n", mprisWatcher.playerBusName, audioFilePath)
				}
				if mprisWatcher.CallBack != nil {
					mprisWatcher.CallBack.Paused(mprisWatcher.playerBusName, audioFilePath, mprisWatcher.lyric)
				}
			default:
				if withLog {
					log.Printf("[WARN] Unknown playback status: %s\n", status)
				}
			}
		}
	}
}

func (mprisWatcher *MprisWatcher) getAllProperties() error {
	// 获取当前播放的音频播放器的 Object
	if mprisWatcher.playerBusName == "" {
		return fmt.Errorf("no player bus name set")
	}

	// 获取当前播放器的 Object
	obj := mprisWatcher.conn.Object(mprisWatcher.playerBusName, "/org/mpris/MediaPlayer2")

	// 调用 "GetAll" 方法获取所有的属性
	var properties map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.Properties.GetAll", 0, "org.mpris.MediaPlayer2.Player").Store(&properties)
	if err != nil {
		return fmt.Errorf("failed to get all properties: %v", err)
	}
	return nil
}

// 获取音频文件时长（单位：微秒）
func (mprisWatcher *MprisWatcher) getSongDuration(audioFilePath string) (uint64, error) {
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
func (mprisWatcher *MprisWatcher) getPosition() (uint64, error) {
	obj := mprisWatcher.conn.Object(mprisWatcher.playerBusName, "/org/mpris/MediaPlayer2")
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
