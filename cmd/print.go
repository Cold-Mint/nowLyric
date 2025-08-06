package cmd

import (
	"nowlyric/lyrics"
	"strconv"

	"github.com/spf13/cobra"
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
		var outputPath = cmd.Flag("outputPath").Value.String()
		delayVal, err := strconv.ParseUint(delayStr, 10, 32)
		if err != nil {
			delayVal = 100
		}
		mprisWatcher := &lyrics.MprisWatcher{}
		err = mprisWatcher.ConnectSessionBus(withLog)
		if err != nil {
			return
		}
		mprisWatcher.CallBack = &lyrics.LyricCallback{OnlyTranslation: onlyTranslation, OutputPath: outputPath}
		go mprisWatcher.SynchronizedLyrics(withLog, uint32(delayVal))
		println("The lyrics monitoring process is ready. It will take effect when you start playing music or switch to the next song.")
		mprisWatcher.MonitorAudioChanges(withLog)
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
	printCmd.Flags().StringP("outputPath", "o", "", "Synchronize and output the latest line of lyrics to the text file.")
	printCmd.Flags().Uint32P("delay", "d", 100, "The delay for synchronizing lyrics, measured in milliseconds.")
	printCmd.Flags().BoolP("withLog", "l", false, "Whether to output logs.")
	printCmd.Flags().BoolP("onlyTranslation", "t", false, "Only display the translation.")
}
