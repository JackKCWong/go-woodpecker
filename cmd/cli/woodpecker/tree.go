package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker"
	"github.com/JackKCWong/go-woodpecker/api"
	"github.com/JackKCWong/go-woodpecker/cmd/cli/config"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/JackKCWong/go-woodpecker/spi/impl/maven"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"text/tabwriter"
)

var (
	cCritical     = color.New(color.FgRed).SprintFunc()
	cHigh         = color.New(color.FgHiRed).SprintFunc()
	cMedium       = color.New(color.FgHiYellow).SprintFunc()
	cLow          = color.New(color.FgYellow).SprintFunc()
	cNode         = color.New(color.FgWhite).SprintFunc()
	cShouldUpdate = color.New(color.FgHiWhite).Add(color.Underline, color.Bold).SprintFunc()
)

var vulTreeCmd = &cobra.Command{
	Use:     "tree",
	Short:   "Print dependency tree with CVEs",
	Aliases: []string{"ls"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return config.ReadConfigFile()
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		summaryMode, _ := cmd.Flags().GetBool("summary")

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		depMgr := maven.New(path.Join(wd, "pom.xml"),
			maven.Opts{
				Output:               config.NewProgressOutput(),
				DependencyCheckProps: viper.GetStringSlice("maven.dependency-check"),
			})

		ossIndexClient, err := config.NewOSSIndexClient()
		if err != nil {
			return err
		}

		wp := woodpecker.Woodpecker{DepMgr: depMgr, OSSIndex: ossIndexClient}
		tree, err := wp.Tree(woodpecker.TreeOpts{Opts: woodpecker.Opts{
			Verbose: viper.GetBool("verbose"),
		}})

		if err != nil {
			return err
		}

		fmt.Fprintln(os.Stdout)
		if summaryMode {
			printSummary(os.Stdout, tree)
		} else {
			printTree(os.Stdout, tree)
		}

		return nil
	},
}

func printSummary(w io.Writer, tree api.DependencyTree) {
	util.Printfln(w, "%s", tree.Root().ID)
	for i, n := range tree.Nodes() {
		if n.Scope == "test" {
			continue
		}

		if n.Depth == 1 {
			subtree, _ := tree.Subtree(i, n.ID)
			nColor := cNode
			prefix := ""
			suffix := ""

			if subtree.VulnerabilityCount() > 0 {
				nColor = cShouldUpdate
				suffix = "\t\t(" + strconv.Itoa(subtree.VulnerabilityCount()) + " vulnerabilities)"
			}

			prefix = "+ "

			padding := "  "
			util.Printfln(w, "%s%s%s%s", padding, prefix, nColor(n.ID), suffix)
			printVulnerabilities(w, subtree.AllVulnerabilities(), padding)
		}
	}
}

func printTree(w io.Writer, tree api.DependencyTree) {
	for i, n := range tree.Nodes() {
		if n.Scope == "test" {
			continue
		}

		padding := strings.Repeat("  ", n.Depth)
		nColor := cNode
		prefix := ""
		suffix := ""

		if n.Depth == 1 {
			subtree, _ := tree.Subtree(i, n.ID)
			if subtree.VulnerabilityCount() > 0 {
				nColor = cShouldUpdate
				suffix = "\t\t(" + strconv.Itoa(subtree.VulnerabilityCount()) + " vulnerabilities)"
			}
		}

		switch {
		case n.Depth == 1:
			prefix = "+ "
		case n.Depth > 1:
			prefix = "- "
		}

		util.Printfln(w, "%s%s%s%s", padding, prefix, nColor(n.ID), suffix)
		if len(n.Vulnerabilities) > 0 {
			printVulnerabilities(w, n.Vulnerabilities, padding)
		}
	}
}

func printVulnerabilities(w io.Writer, vuls []api.Vulnerability, padding string) {
	tw := new(tabwriter.Writer)
	tw.Init(w, 10, 0, 2, ' ', 0)

	for i, v := range vuls {
		var vColor func(...interface{}) string
		switch {
		case v.Severity() == "Critical":
			vColor = cCritical
		case v.Severity() == "High":
			vColor = cHigh
		case v.Severity() == "Medium":
			vColor = cMedium
		default:
			vColor = cLow
		}

		util.Printfln(tw, "%s   %d\t%s\t%s\t%.1f\t%s", padding, i+1,
			vColor(v.Cve), vColor(v.Severity()), v.CvssScore, v.NVDUrl())
	}
	tw.Flush()
}

func init() {
	vulTreeCmd.Flags().BoolP("summary", "s", false, "print summary only")
}
