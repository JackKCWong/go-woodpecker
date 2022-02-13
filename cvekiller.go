package woodpecker

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"os"
	"strings"
	"time"
)

type KillOpts struct {
	Opts
	SendPR bool
}

func (w Woodpecker) Kill(args []string, opts KillOpts) error {
	if multiModules, err := w.DepMgr.IsMultiModules(); err != nil {
		return fmt.Errorf("failed to check if multi-modules: %w", err)
	} else if multiModules {
		return fmt.Errorf("kill does not work on the parent project. Please run it on the child project")
	}

	tree, err := w.DepMgr.DependencyTree()
	if err != nil {
		return err
	}

	cveID := args[0]

	subtreeToUpdate, found := tree.FirstChildWithCVE(cveID)
	if !found {
		return fmt.Errorf("CVE %s not found in the dependency tree", cveID)
	}

	originalPackageID := subtreeToUpdate.Root().ID
	newPackageID := ""

	newBrachName := strings.ReplaceAll(fmt.Sprintf("%s/%s/%s", opts.BranchNamePrefix, originalPackageID, cveID), ":", "/")
	err = w.GitClient.Branch(newBrachName)
	if err != nil {
		return err
	}

	for subtreeToUpdate, found = subtreeToUpdate.FirstChildWithCVE(cveID); found; subtreeToUpdate, found = subtreeToUpdate.FirstChildWithCVE(cveID) {
		util.Printfln(os.Stdout, "%s found in %s, upgrading...", cveID, subtreeToUpdate.Root().ID)
		lastPackageID := subtreeToUpdate.Root().ID
		newPackageID, err = w.DepMgr.UpdateDependency(subtreeToUpdate.Root())
		if err != nil {
			return fmt.Errorf("failed to update dependency %s: %w", subtreeToUpdate.Root().ID, err)
		}
		if lastPackageID == newPackageID {
			util.Printfln(os.Stdout, "already the latest version: %s, exiting...", newPackageID)
			return fmt.Errorf("no version available without %s", cveID)
		}

		util.Printfln(os.Stdout, "upgraded to %s", newPackageID)

		subtreeToUpdate, err = w.DepMgr.DependencyTree()
		if err != nil {
			return fmt.Errorf("failed to get dependency tree: %w", err)
		}
	}

	util.Printfln(os.Stdout, "%s is killed.", cveID)
	util.Printfln(os.Stdout, "start verifying...")
	result, err := w.DepMgr.Verify()
	if !result.Passed {
		if err == nil {
			err = fmt.Errorf("unknown error")
		}

		return fmt.Errorf("verification failed: %w\n%s", err, result.Summary)
	}

	var verificationResult string
	if result.Summary == "" {
		verificationResult = "build passed but you don't seem to have any test! good luck!"
		util.Printfln(os.Stdout, verificationResult)
	} else {
		verificationResult = fmt.Sprintf("verification passed: \n%s", result.Summary)
		util.Printfln(os.Stdout, verificationResult)
	}

	err = w.DepMgr.StageUpdate()
	if err != nil {
		return fmt.Errorf("failed to apply change: %w", err)
	}

	commitMessage := "removing " + cveID + " in " + originalPackageID
	hash, err := w.GitClient.Commit(commitMessage)
	if err != nil {
		return err
	}

	util.Printfln(os.Stdout, "commited %s", hash)

	if opts.SendPR {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel()

		err := w.GitClient.Push(ctx)
		if err != nil {
			return err
		}

		origin, err := w.GitClient.Origin()
		if err != nil {
			return err
		}

		prComment := fmt.Sprintf("### updated to %s, verification result:\n> %s", newPackageID, verificationResult)
		subtreeToUpdate, found := tree.FirstChildWithCVE(cveID)
		if found {
			cve, _ := subtreeToUpdate.FindCVE(cveID)
			prComment = fmt.Sprintf("%s\n\n### CVE details\n%s\n>%s",
				prComment, formatCVESummary(cve), cve.Description)
		} else {
			util.Printfln(os.Stderr, "weired...CVE %s not found in the original dependency tree...", cveID)
		}

		pullRequestURL, err := w.GitServer.CreatePullRequest(ctx,
			origin, newBrachName, "master",
			commitMessage,
			prComment)

		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "Pull request created: %s", pullRequestURL)
	}

	return nil
}

func formatCVESummary(cve api.Vulnerability) string {
	return fmt.Sprintf("[%s](%s) - %s %.1f", cve.Cve, cve.NVDUrl(), cve.Severity, cve.CvssScore)
}
