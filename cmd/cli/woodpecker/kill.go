package main

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

var killCmd = &cobra.Command{
	Use:   "kill cve_id",
	Short: "Update dependency version until the given CVE disappears",
	Long:  "Update dependency version until the given CVE disappears and number of critical vulnerabilities drops",
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return readViperConf()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		gitClient, err := newGitClient()
		if err != nil {
			return err
		}

		err = gitClient.CreateBranch(viper.GetString("branch-name"))
		if err != nil {
			return err
		}

		depMgr := maven.NewRunner("pom.xml",
			maven.Opts{
				Output:               newProgressOutput(),
				DependencyCheckProps: viper.GetStringMapString("maven.dependency-check"),
			})

		tree, err := depMgr.DependencyTree()
		if err != nil {
			return err
		}

		cveID := args[0]
		subtree, found := tree.FindCVE(cveID)
		originalPackageID := subtree.Root().ID
		for found {
			util.Printfln(os.Stdout, "%s found in %s, upgrading...", cveID, subtree.Root().ID)
			err := depMgr.UpdateDependency(subtree.Root().ID)
			if err != nil {
				return fmt.Errorf("failed to update dependency %s: %w", subtree.Root().ID, err)
			}

			newTree, err := depMgr.DependencyTree()
			if err != nil {
				return fmt.Errorf("failed to get dependency tree: %w", err)
			}

			subtree, found = newTree.FindCVE(cveID)
		}

		util.Printfln(os.Stdout, "%s is killed.", cveID)
		err = depMgr.StageUpdate()
		if err != nil {
			return fmt.Errorf("failed to apply change: %w", err)
		}

		hash, err := gitClient.Commit("removing " + cveID + " in " + originalPackageID)
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "commited %s", hash)

		if viper.GetBool("send-pr") {
			err := gitClient.Push()
			if err != nil {
				return err
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			gitHub, err := newGitHubClient()
			if err != nil {
				return err
			}

			origin, err := gitClient.Origin()
			if err != nil {
				return err
			}

			pullRequestURL, err := gitHub.CreatePullRequest(ctx, origin, viper.GetString("branch-name"), "master")
			if err != nil {
				return err
			}

			util.Printfln(os.Stdout, "Pull request created: %s", pullRequestURL)
		}

		return nil
	},
}

func init() {
	killCmd.Flags().Bool("send-pr", false, "Create a PR if successful.")
}
