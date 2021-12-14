package main

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/gitop"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var digCmd = &cobra.Command{
	Use:   "dig",
	Short: "dig out a dependency which can be upgraded to reduce vulnerabilities",
	RunE: func(cmd *cobra.Command, args []string) error {
		updater := maven.NewUpdater(
			"pom.xml",
			maven.UpdaterOpts{Verbose: true},
		)

		depTree, err := updater.DependencyTree()
		if err != nil {
			return err
		}

		vulnerable, found := depTree.MostVulnerable()
		if !found {
			fmt.Println("Congradulation! Your project has no CVE.")
			return nil
		}

		root, _ := depTree.Find(vulnerable.Root().ID)
		util.Printfln(os.Stdout, "updating dependencies %s with %d vulnerabilities", root.ID, len(vulnerable.AllVulnerabilities()))
		err = updater.UpdateDependency(root.ID)
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

		gitClient := gitop.GitClient{RepoDir: "./"}

		_, err = gitClient.CommitAndPush("woodpecker-autoupdate", "update "+vulnerable.Root().ID)
		if err != nil {
			return err
		}

		origin, err := gitClient.Origin()
		if err != nil {
			return err
		}

		gitHub := gitop.GitHub{AccessToken: os.Getenv("WOODPECKER_GITHUB_ACCESSTOKEN")}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err = gitHub.CreatePullRequest(ctx, origin, "woodpecker-autoupdate", "master")
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {

}
