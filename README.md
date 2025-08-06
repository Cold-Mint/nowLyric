# nowLyric

### This program is applicable to the Linux system. It has not yet been tested on other systems.

The console program is capable of listening for events when the system plays music. And when the music is playing, load the lyrics from the local lrc file.

### Usage:

nowlyric print [flags]

Flags:
1. -d, --delay uint32        The delay for synchronizing lyrics, measured in milliseconds. (default 100)
2. -h, --help                help for print.
3. -t, --onlyTranslation     Only display the translation.
4. -o, --outputPath string   Synchronize and output the latest line of lyrics to the text file.
5. -l, --withLog             Whether to output logs.

### 此程序适用于Linux系统。尚未在其他系统进行测试。

控制台程序，能够监听系统播放音乐的事件。并在音乐播放时，从本地lrc文件加载歌词。

使用

nowlyric print [flags]

Flags:
1. -d, --delay uint32        同步歌词的延迟，以毫秒为单位。100(默认)
2. -h, --help                打印帮助。
3. -t, --onlyTranslation     只显示翻译。
4. -o, --outputPath string   同步并输出最新一行歌词到文本文件。
5. -l, --withLog             是否输出日志。