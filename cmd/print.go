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
		var richText = cmd.Flag("richText").Value.String() == "true"
		var supportExecute = cmd.Flag("supportExecute").Value.String() == "true"
		var playedTextColor = cmd.Flag("playedTextColor").Value.String()
		var unplayedTextColor = cmd.Flag("unplayedTextColor").Value.String()
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
		MPrisListener.CallBack = &lyrics.LyricCallback{OnlyTranslation: onlyTranslation, OutputPath: outputPath, RichText: richText, SupportExecute: supportExecute, PlayedTextColor: playedTextColor, UnplayedTextColor: unplayedTextColor, Offset: offset, WithLog: withLog}
		go MPrisListener.SynchronizedLyrics(withLog, uint32(delayVal))
		println("The lyrics monitoring process is ready. It will take effect when you start playing music or switch to the next song.")
		MPrisListener.WatchPlayerEvents(withLog)
	},
}

func init() {
	rootCmd.AddCommand(printCmd)
	printCmd.Flags().StringP("outputPath", "o", "", "Synchronize and output the latest line of lyrics to the text file.")
	printCmd.Flags().Uint32P("delay", "d", 100, "The delay for synchronizing lyrics, measured in milliseconds.")
	printCmd.Flags().BoolP("withLog", "l", false, "Whether to output logs.")
	printCmd.Flags().BoolP("onlyTranslation", "t", false, "Only display the translation.")
	printCmd.Flags().BoolP("richText", "r", false, "Use colored text. For example: <span foreground='color'>text</span>.")
	printCmd.Flags().BoolP("supportExecute", "e", false, "richText needs to be enabled.Support for Executor-Gnome Shell Extension color font format.After enabling it, <executor.markup.true> will be added before the output.")
	printCmd.Flags().StringP("playedTextColor", "p", "#FFFFFF", "richText needs to be enabled.Define the text color of the played part, with the default being #FFFFFF.")
	printCmd.Flags().StringP("unplayedTextColor", "u", "#FFFFFF", "richText needs to be enabled.Define the text color for the unplayed part, with the default being #FFFFFF.")
	printCmd.Flags().Float64("offset", 0.05, "The offset used for the playback progress. Between 0 and 1. For example: This line of lyrics has actually been played by 50%. The program will add an offset to generate the rendered text. If the offset is 0.1, then 50%+0.1 (10%) =60%.Default 0.05 (%5).")
}
