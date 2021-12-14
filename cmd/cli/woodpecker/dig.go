package main

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/api"
	"github.com/JackKCWong/go-woodpecker/internal/spi/gitop"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var digCmd = &cobra.Command{
	Use:   "dig [package_id]",
	Short: "dig out a dependency which can be upgraded to reduce vulnerabilities. package_id is in the format of groupId:artifactId:version",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		util.Printfln(os.Stdout, "--------------------------------------------")
		util.Printfln(os.Stdout, "updating dependencies %s with %d vulnerabilities", target.Root().ID, target.VulnerabilityCount())
		util.Printfln(os.Stdout, "--------------------------------------------")
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

		gitDir, err := gitop.FindGitDir("./")
		if err != nil {
			return err
		}

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

		gitHub := gitop.GitHub{AccessToken: os.Getenv("WOODPECKER_GITHUB_ACCESSTOKEN")}

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

}
