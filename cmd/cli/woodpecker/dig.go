package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker"
	"github.com/JackKCWong/go-woodpecker/cmd/cli/config"
	"github.com/JackKCWong/go-woodpecker/spi/impl/maven"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var digCmd = &cobra.Command{
	Use:   "dig",
	Short: "dig out a dependency with Critical or High CVE and try to fix it",
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

		depMgr := maven.New(
			"pom.xml",
			maven.Opts{
				Output:               config.NewProgressOutput(),
				DependencyCheckProps: viper.GetStringSlice("maven.dependency-check"),
			},
		)

		wp := woodpecker.Woodpecker{
			GitClient: gitClient,
			GitServer: gitHub,
			DepMgr:    depMgr,
		}

		depTree, err := depMgr.DependencyTree()
		if err != nil {
			return err
		}

		target, found := depTree.CriticalOrHigh()
		if !found {
			fmt.Println("Congratulations! Your project has no Critical or High CVE.")
			return nil
		}

		return wp.Kill([]string{target.ID}, opts)
	},
}
