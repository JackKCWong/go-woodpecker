package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
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

		return nil
	},
}

func init() {
	killCmd.Flags().Bool("no-pr", false, "do not creat a PR. useful for debug")
}
