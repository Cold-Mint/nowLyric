# nowLyric

### This program is applicable to the Linux system. It has not yet been tested on other systems.

The console program is capable of listening for events when the system plays music. And when the music is playing, load
the lyrics from the local lrc file.

### Usage:

nowlyric print [flags]

Flags:

- -d, --delay uint32 The delay for synchronizing lyrics, measured in milliseconds. (default 100)
- -h, --help help for print
- --offset float The offset used for the playback progress. Between 0 and 1. For example: This line of lyrics has
  actually been played by 50%. The program will add an offset to generate the rendered text. If the offset is 0.1, then
  50%+0.1 (10%) =60%.Default 0.05 (%5). (default 0.05)
- -t, --onlyTranslation Only display the translation.
- -s, --sharedMemory Create a memory area on your device that can be shared by multiple processes using shared memory.
  Note: To use the nowlyric read command, this flag needs to be enabled.
-
    - -p, --playedTextColor string richText needs to be enabled.Define the text color of the played part, with the
      default
      being #FFFFFF. (default "#FFFFFF")
- -r, --richText Use colored text. For example: <span foreground='color'>text</span>.
- -e, --supportExecute richText needs to be enabled.Support for Executor-Gnome Shell Extension color font format.After
  enabling it, <executor.markup.true> will be added before the output.
- -u, --unplayedTextColor string richText needs to be enabled.Define the text color for the unplayed part, with the
  default being #FFFFFF. (default "#FFFFFF")
- -l, --withLog Whether to output logs.

nowlyric print read

Read the lyrics that are playing.

### 此程序适用于Linux系统。尚未在其他系统进行测试。

控制台程序，能够监听系统播放音乐的事件。并在音乐播放时，从本地lrc文件加载歌词。

使用

nowlyric print [flags]

Flags:

- -d, --delay uint32 同步歌词的延迟，以毫秒为单位。100(默认)
- -h, --help 打印帮助
- --offset float
  用于播放进度的偏移量。在0到1之间。这句歌词实际上已经播放50%。该程序将添加一个偏移量来生成渲染文本。例如：偏移量为0.1，则50%+0.1(
  10%)=60%。默认值0.05（%5）。
- -t, --onlyTranslation 只显示翻译。
- -s, --sharedMemory 在您的设备上创建一个可以由使用共享内存的多个进程共享的内存区域。注意：要使用nowlyric read命令，需要启用此标志。
- -p, --playedTextColor string richText需要被启用。定义已播放的部分文本颜色，默认为#FFFFFF。
- -r, --richText 使用彩色文本。例如：<span foreground='color'>text</span>。
- -e, --supportExecute richText需要被启用。支持Executor-Gnome Shell扩展颜色字体格式。启用后，使用<executor.markup。True >
  将在输出前添加。
- -u, --unplayedTextColor string richText需要被启用。定义未播放部分的文本颜色，默认为#FFFFFF。
- -l, --withLog 是否输出日志。

nowlyric print read

读取正在播放的歌词。