package lyrics

// MusicEventCallback Music event callback
// 音乐事件回调
type MusicEventCallback interface {

	// Play
	//播放音频时
	Play(playerBusName string, audioFilePath string, lyric *Lyric)

	// Stop
	//停止播放音频时
	Stop(playerBusName string, audioFilePath string, lyric *Lyric)

	// Paused
	// 暂停播放音频时
	Paused(playerBusName string, audioFilePath string, lyric *Lyric)

	// UpdateLyric
	// 当需要更新歌词时
	UpdateLyric(playerBusName string, line string, progress float64, lyric *Lyric)
}
