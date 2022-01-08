package woodpecker

import (
	"context"
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

type KillOpts struct {
	Opts
	SendPR bool
}

func (w Woodpecker) Kill(args []string, opts KillOpts) error {
	pomPath := "pom.xml"
	pomXml, err := ioutil.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("failed to read pom.xml in current wd: %w", err)
	}

	if strings.Contains(string(pomXml), "<packaging>pom</packaging>") {
		return fmt.Errorf("kill does not work on the parent project. Please run it on the child project")
	}

	tree, err := w.DepMgr.DependencyTree()
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

	newBrachName := strings.ReplaceAll(fmt.Sprintf("%s/%s/%s", opts.BranchNamePrefix, originalPackageID, cveID), ":", "/")
	err = w.GitClient.Branch(newBrachName)
	if err != nil {
		return err
	}

	for depWithCVE, found := subtree.FindCVE(cveID); found; depWithCVE, found = subtree.FindCVE(cveID) {
		util.Printfln(os.Stdout, "%s found in %s, upgrading...", cveID, depWithCVE.Root().ID)
		lastPackageID := depWithCVE.Root().ID
		newPackageID, err = w.DepMgr.UpdateDependency(depWithCVE.Root())
		if err != nil {
			return fmt.Errorf("failed to update dependency %s: %w", depWithCVE.Root().ID, err)
		}
		if lastPackageID == newPackageID {
			util.Printfln(os.Stdout, "already the latest version: %s, exiting...", newPackageID)
			return fmt.Errorf("no version available without %s", cveID)
		}

		util.Printfln(os.Stdout, "upgraded to %s", newPackageID)

		subtree, err = w.DepMgr.DependencyTree()
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
		verificationResult = "verification passed but you don't seem to have any test! good luck!"
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

		pullRequestURL, err := w.GitServer.CreatePullRequest(ctx,
			origin, newBrachName, "master",
			commitMessage,
			fmt.Sprintf("update from %s to %s, result:\n%s", originalPackageID, newPackageID, verificationResult))

		if err != nil {
			return err
		}

		util.Printfln(os.Stdout, "Pull request created: %s", pullRequestURL)
	}

	return nil
}
