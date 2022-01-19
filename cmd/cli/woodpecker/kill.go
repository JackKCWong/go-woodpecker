package main

import (
	"github.com/JackKCWong/go-woodpecker"
	"github.com/JackKCWong/go-woodpecker/cmd/cli/config"
	"github.com/JackKCWong/go-woodpecker/spi/impl/maven"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path"
)

var killCmd = &cobra.Command{
	Use:   "kill cve_id",
	Short: "Update dependency version until the given CVE disappears",
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.ReadConfigFile()
	},
	RunE: func(cmd *cobra.Command, args []string) error {

		gitClient, err := config.NewGitClient()
		if err != nil {
			return err
		}

		gitHub, err := config.NewGitHub()
		if err != nil {
			return err
		}

		opts := woodpecker.KillOpts{
			Opts: woodpecker.Opts{
				BranchNamePrefix: viper.GetString("branch-name"),
			},
			SendPR: viper.GetBool("send-pr"),
		}

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		wp := woodpecker.Woodpecker{
			GitClient: gitClient,
			GitServer: gitHub,
			DepMgr: maven.New(path.Join(wd, "pom.xml"),
				maven.Opts{
					Output:               config.NewProgressOutput(),
					DependencyCheckProps: viper.GetStringSlice("maven.dependency-check"),
				}),
		}

		return wp.Kill(args, opts)
	},
}

func init() {
	killCmd.Flags().Bool("send-pr", false, "Create a PR if successful.")
}
