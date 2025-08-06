package lyrics

type MprisCallBack interface {
	//播放某个音乐时（路径可能为空）
	Play(playerBusName string, audioFilePath string, lyric *Lyric)

	//停止播放某个音乐时（路径可能为空）
	Stop(playerBusName string, audioFilePath string, lyric *Lyric)

	//暂停播放某个音乐时（路径可能为空）
	Paused(playerBusName string, audioFilePath string, lyric *Lyric)

	UpdateLyric(playerBusName string, line string, progress float64, lyric *Lyric)
}
