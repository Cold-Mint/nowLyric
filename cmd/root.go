package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "nowlyric",
	Short: "Get the lyrics corresponding to the currently playing music.",
	Long: `Get the lyrics corresponding to the currently playing music. 
The song files and lyrics files must be placed in the same-level directory and have matching file names. 
For example: a.lac matches a.lrc.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
