package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "woodpecker",
	Short: "A collections of tools to help developer to deal with vulnerabiliies",
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(vulTreeCmd)
}

func initConfig() {

}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		util.Printfln(os.Stderr, "exit with error: %q", err)
	}
}
