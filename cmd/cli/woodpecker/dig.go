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
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error
				util.Printfln(os.Stdout, "config not found, continue...")
			} else {
				// Config file was found but another error was produced
				return fmt.Errorf("failed to read config file: %w", err)
			}
		}

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

		gitDir, err := gitop.FindGitDir("./")
		if err != nil {
			return err
		}

		updater := maven.NewUpdater(
			"pom.xml",
			maven.UpdaterOpts{Verbose: true},
		)

		depTree, err := updater.DependencyTree()
		if err != nil {
			return err
		}

		var target api.DependencyTree
		var found bool
		if len(args) > 0 {
			packageID := args[0]
			target, found = depTree.Subtree(packageID)
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
		err = updater.UpdateDependency(target.Root().ID)
		if err != nil {
			return err
		}

		err = updater.Verify()
		if err != nil {
			return err
		}

		err = updater.StageUpdate()
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "opening git repo")

		gitClient := gitop.GitClient{RepoDir: gitDir}

		origin, err := gitClient.Origin()
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "pushing changes to  %s", origin)
		_, err = gitClient.CommitAndPush("woodpecker-autoupdate", "update "+target.Root().ID)
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
		err = gitHub.CreatePullRequest(ctx, origin, "woodpecker-autoupdate", "master")
		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "pull request created")
		return nil
	},
}

func init() {
	digCmd.Flags().String("github-url", "", "github api url. e.g. https://api.github.com")
	viper.BindPFlag("github.url", digCmd.Flags().Lookup("github-url"))

	digCmd.Flags().String("github-accesstoken", "", "github access token")
	viper.BindPFlag("github.accesstoken", digCmd.Flags().Lookup("github-accesstoken"))
}
