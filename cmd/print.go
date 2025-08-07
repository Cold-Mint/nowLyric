package cmd

/*
#cgo LDFLAGS: -lrt
#include <fcntl.h>
#include <sys/mman.h>
#include <unistd.h>
#include <stdlib.h>   // 让 cgo 找到 free
*/
import "C"
import (
	"fmt"
	"nowlyric/lyrics"
	"strconv"
	"unsafe"

	"github.com/spf13/cobra"
)

const (
	name        = "/my_go_shm"
	shmNameCStr = "/my_go_shm\x00" // 以 \0 结尾的 C 字符串
)

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print",
	Short: "Get the lyrics currently playing and print them out.",
	Long:  `Get the lyrics currently playing and print them out.`,
	Run: func(cmd *cobra.Command, args []string) {
		var withLog = cmd.Flag("withLog").Value.String() == "true"
		var delayStr = cmd.Flag("delay").Value.String()
		var onlyTranslation = cmd.Flag("onlyTranslation").Value.String() == "true"
		var richText = cmd.Flag("richText").Value.String() == "true"
		var supportExecute = cmd.Flag("supportExecute").Value.String() == "true"
		var playedTextColor = cmd.Flag("playedTextColor").Value.String()
		var unplayedTextColor = cmd.Flag("unplayedTextColor").Value.String()
		var defaultContent = cmd.Flag("defaultContent").Value.String()
		var sharedMemory = cmd.Flag("sharedMemory").Value.String() == "true"
		//指针默认为null，只有使用sharedMemory才为其赋值
		var ptr unsafe.Pointer
		var mmapOK = false
		if sharedMemory {
			cName := C.CString(shmNameCStr)
			defer C.free(unsafe.Pointer(cName))
			fd, err := C.shm_open(cName, C.O_CREAT|C.O_RDWR, 0666)
			if fd < 0 {
				panic(fmt.Sprintf("shm_open failed: %v", err))
			}
			defer C.close(fd)
			if _, err := C.ftruncate(fd, lyrics.Size); err != nil {
				panic(fmt.Sprintf("ftruncate failed: %v", err))
			}
			ptr, err = C.mmap(nil, lyrics.Size, C.PROT_READ|C.PROT_WRITE, C.MAP_SHARED, fd, 0)
			if ptr == C.MAP_FAILED {
				panic(fmt.Sprintf("mmap failed: %v", err))
			}
			mmapOK = true
			defer C.munmap(ptr, lyrics.Size)
			lyrics.WriteCString(ptr, defaultContent, lyrics.Size)
		}
		raw := cmd.Flag("offset").Value.String()
		offset, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			offset = 0.05
		}
		if playedTextColor == "" {
			playedTextColor = "#FFFFFF"
		}
		if unplayedTextColor == "" {
			unplayedTextColor = "#FFFFFF"
		}
		delayVal, err := strconv.ParseUint(delayStr, 10, 32)
		if err != nil {
			delayVal = 100
		}
		MPrisListener := &lyrics.MPrisListener{}
		err = MPrisListener.ConnectSessionBus(withLog)
		if err != nil {
			return
		}
		MPrisListener.CallBack = &lyrics.LyricCallback{OnlyTranslation: onlyTranslation, RichText: richText, SupportExecute: supportExecute, PlayedTextColor: playedTextColor, UnplayedTextColor: unplayedTextColor, Offset: offset, WithLog: withLog, MmapOK: mmapOK, Ptr: ptr}
		go MPrisListener.SynchronizedLyrics(withLog, uint32(delayVal))
		println("The lyrics monitoring process is ready. It will take effect when you start playing music or switch to the next song.")
		MPrisListener.WatchPlayerEvents(withLog)
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
	printCmd.Flags().StringP("defaultContent", "c", "", "The outputPath must not be empty.The content of the file written by default when the program starts.")
	printCmd.Flags().Uint32P("delay", "d", 100, "The delay for synchronizing lyrics, measured in milliseconds.")
	printCmd.Flags().BoolP("withLog", "l", false, "Whether to output logs.")
	printCmd.Flags().BoolP("onlyTranslation", "t", false, "Only display the translation.")
	printCmd.Flags().BoolP("richText", "r", false, "Use colored text. For example: <span foreground='color'>text</span>.")
	printCmd.Flags().BoolP("supportExecute", "e", false, "richText needs to be enabled.Support for Executor-Gnome Shell Extension color font format.After enabling it, <executor.markup.true> will be added before the output.")
	printCmd.Flags().StringP("playedTextColor", "p", "#FFFFFF", "richText needs to be enabled.Define the text color of the played part, with the default being #FFFFFF.")
	printCmd.Flags().StringP("unplayedTextColor", "u", "#FFFFFF", "richText needs to be enabled.Define the text color for the unplayed part, with the default being #FFFFFF.")
	printCmd.Flags().BoolP("sharedMemory", "s", false, "Create a memory area on your device that can be shared by multiple processes using shared memory. Note: To use the nowlyric read command, this flag needs to be enabled.")
	printCmd.Flags().Float64("offset", 0.05, "The offset used for the playback progress. Between 0 and 1. For example: This line of lyrics has actually been played by 50%. The program will add an offset to generate the rendered text. If the offset is 0.1, then 50%+0.1 (10%) =60%.Default 0.05 (%5).")
}
