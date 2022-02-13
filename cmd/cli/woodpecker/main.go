package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().Bool("no-progress", false, "suppress progress spinner")

	rootCmd.AddCommand(
		vulTreeCmd,
		digCmd,
		killCmd,
	)

	bindCmdOptsToViperConf(
		rootCmd.PersistentFlags(),
		vulTreeCmd.Flags(),
		killCmd.Flags(),
		digCmd.Flags(),
	)
	// defaults
	viper.SetDefault("verbose", false)
	viper.SetDefault("no-progress", false)
	viper.SetDefault("branch-name", "woodpecker")

	viper.SetConfigName(".woodpecker")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(os.ExpandEnv("$HOME/"))
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("WOODPECKER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv()
}

// bindCmdOptsToViperConf replace '-' with '.' before binding so it can bind to nested properties more naturally
// eg. foo-bar is bound to foo.bar
func bindCmdOptsToViperConf(flags ...*pflag.FlagSet) {
	for _, f := range flags {
		_ = viper.BindPFlags(f)
		f.VisitAll(func(f *pflag.Flag) {
			viper.BindPFlag(strings.Replace(f.Name, "-", ".", 1), f)
		})
	}
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		util.Printfln(os.Stderr, "exit with error: %q", err)
	}
}
