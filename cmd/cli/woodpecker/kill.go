package main

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/cmd/cli/config"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/JackKCWong/go-woodpecker/spi/impl/maven"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var killCmd = &cobra.Command{
	Use:   "kill cve_id",
	Short: "Update dependency version until the given CVE disappears",
	Args:  cobra.MinimumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.ReadConfigFile()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		pomPath := "pom.xml"
		pomXml, err := ioutil.ReadFile(pomPath)
		if err != nil {
			return fmt.Errorf("failed to read pom.xml in current wd: %w", err)
		}

		if strings.Contains(string(pomXml), "<packaging>pom</packaging>") {
			return fmt.Errorf("kill does not work on the parent project. Please run it on the child project")
		}

		gitClient, err := config.NewGitClient()
		if err != nil {
			return err
		}

		newBrachName := viper.GetString("branch-name")
		err = gitClient.Branch(newBrachName)
		if err != nil {
			return err
		}

		depMgr := maven.New("pom.xml",
			maven.Opts{
				Output:               config.NewProgressOutput(),
				DependencyCheckProps: viper.GetStringSlice("maven.dependency-check"),
			})

		tree, err := depMgr.DependencyTree()
		if err != nil {
			return err
		}

		cveID := args[0]

		subtree, found := tree.FindCVE(cveID)
		if !found {
			return fmt.Errorf("CVE %s not found in the dependency tree", cveID)
		}

		originalPackageID := subtree.Root().ID
		newPackageID := ""

		for depWithCVE, found := subtree.FindCVE(cveID); found; depWithCVE, found = subtree.FindCVE(cveID) {
			util.Printfln(os.Stdout, "%s found in %s, upgrading...", cveID, depWithCVE.Root().ID)
			lastPackageID := depWithCVE.Root().ID
			newPackageID, err = depMgr.UpdateDependency(depWithCVE.Root())
			if err != nil {
				return fmt.Errorf("failed to update dependency %s: %w", depWithCVE.Root().ID, err)
			}
			if lastPackageID == newPackageID {
				util.Printfln(os.Stdout, "already the latest version: %s, exiting...", newPackageID)
				return fmt.Errorf("no version available without %s", cveID)
			}

			util.Printfln(os.Stdout, "upgraded to %s", newPackageID)

			subtree, err = depMgr.DependencyTree()
			if err != nil {
				return fmt.Errorf("failed to get dependency tree: %w", err)
			}
		}

		util.Printfln(os.Stdout, "%s is killed.", cveID)
		util.Printfln(os.Stdout, "start verifying...")
		result, err := depMgr.Verify()
		if !result.Passed {
			if err == nil {
				err = fmt.Errorf("unknown error")
			}

			return fmt.Errorf("verification failed: %w\n%s", err, result.Summary)
		}

		var verificationResult string
		if result.Summary == "" {
			verificationResult = "verification passed but you don't seem to have any test! good luck!"
			util.Printfln(os.Stdout, verificationResult)
		} else {
			verificationResult = fmt.Sprintf("verification passed: \n%s", result.Summary)
			util.Printfln(os.Stdout, verificationResult)
		}

		err = depMgr.StageUpdate()
		if err != nil {
			return fmt.Errorf("failed to apply change: %w", err)
		}

		commitMessage := "removing " + cveID + " in " + originalPackageID
		hash, err := gitClient.Commit(commitMessage)
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "commited %s", hash)

		if viper.GetBool("send-pr") {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
			defer cancel()

			err := gitClient.Push(ctx)
			if err != nil {
				return err
			}

			gitHub, err := config.NewGitHub()
			if err != nil {
				return err
			}

			origin, err := gitClient.Origin()
			if err != nil {
				return err
			}

			pullRequestURL, err := gitHub.CreatePullRequest(ctx,
				origin, newBrachName, "master",
				commitMessage,
				fmt.Sprintf("update from %s to %s, result:\n%s", originalPackageID, newPackageID, verificationResult))

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
