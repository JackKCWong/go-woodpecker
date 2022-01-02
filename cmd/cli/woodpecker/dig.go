package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/JackKCWong/go-woodpecker/internal/spi/gitop"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/url"
	"os"
	"strings"
	"time"
)

var digCmd = &cobra.Command{
	Use:   "dig [package_id]",
	Short: "dig out a dependency which can be upgraded to reduce vulnerabilities. package_id is in the format of groupId:artifactId:version",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return readViperConf()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		githubUrl := viper.GetString("github.url")
		if githubUrl == "" {
			return errors.New("github.url is not set")
		}

		if !strings.HasSuffix(githubUrl, "/") {
			githubUrl += "/"
		}

		githubURL, err := url.Parse(githubUrl)
		if err != nil {
			return fmt.Errorf("github.url is not a valid url: %w", err)
		}

		githubToken := viper.GetString("github.accesstoken")
		if githubToken == "" {
			return errors.New("github.accesstoken is not set")
		}

		gitClient, err := newGitClient()
		if err != nil {
			return err
		}

		err = gitClient.CreateBranch(viper.GetString("branch-name"))
		if err != nil {
			return fmt.Errorf("failed to create branch: %w", err)
		}

		depMgr := maven.New(
			"pom.xml",
			maven.Opts{
				Output:               newProgressOutput(),
				DependencyCheckProps: viper.GetStringSlice("maven.dependency-check"),
			},
		)

		depTree, err := depMgr.DependencyTree()
		if err != nil {
			return err
		}

		var target api.DependencyTree
		var found bool
		if len(args) > 0 {
			packageID := args[0]
			target, found = depTree.Subtree(0, packageID)
			if !found {
				return fmt.Errorf("package %s not found", packageID)
			}
		} else {
			target, found = depTree.MostVulnerable()
			if !found {
				fmt.Println("Congratulations! Your project has no CVE.")
				return nil
			}
		}

		util.Printfln(os.Stdout, strings.Repeat("-", 80))
		util.Printfln(os.Stdout, "updating dependencies %s with %d vulnerabilities", target.Root().ID, target.VulnerabilityCount())
		util.Printfln(os.Stdout, strings.Repeat("-", 80))
		err = depMgr.UpdateDependency(target.Root().ID)
		if err != nil {
			return err
		}

		r, err := depMgr.Verify()
		if err != nil {
			return err
		}

		if !r.Passed {
			return fmt.Errorf("verification failed: \n%s", r.Summary)
		}

		err = depMgr.StageUpdate()
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "opening git repo")

		origin, err := gitClient.Origin()
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "pushing changes to  %s", origin)
		hash, err := gitClient.Commit("update " + target.Root().ID)
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "committed %s", hash)

		err = gitClient.Push()
		if err != nil {
			return err
		}

		gitHub := gitop.GitHub{
			BaseURL:     githubURL,
			AccessToken: githubToken,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		util.Printfln(os.Stdout, "creating pull request to %s", origin)
		pr, err := gitHub.CreatePullRequest(ctx,
			origin, viper.GetString("branch-name"), "master",
			"upgrading "+target.Root().ID, r.Summary)

		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "pull request created: %s", pr)
		return nil
	},
}
