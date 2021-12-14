package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:          "woodpecker",
	Short:        "A collections of tools to help developer to deal with vulnerabiliies",
	SilenceUsage: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(vulTreeCmd)
	rootCmd.AddCommand(digCmd)
}

func initConfig() {

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		util.Printfln(os.Stderr, "exit with error: %q", err)
	}
}
