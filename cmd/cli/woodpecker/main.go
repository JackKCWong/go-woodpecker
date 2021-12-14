package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:           "woodpecker",
	Short:         "A collections of tools to help developer to deal with vulnerabilities",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.AddCommand(vulTreeCmd)
	rootCmd.AddCommand(digCmd)
}

func initConfig() {
	viper.SetConfigName("conf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.ExpandEnv("$HOME/.woodpecker"))
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("WOODPECKER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		util.Printfln(os.Stderr, "exit with error: %q", err)
	}
}
