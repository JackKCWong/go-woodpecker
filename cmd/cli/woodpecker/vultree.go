package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		noProgress, _ := cmd.Flags().GetBool("no-progress")
		updater := maven.NewUpdater("pom.xml",
			maven.UpdaterOpts{
				Verbose: !noProgress,
			})

		tree, err := updater.DependencyTree()
		if err != nil {
			return err
		}

		fmt.Println()
		for _, n := range tree.Nodes() {
			if n.Scope == "test" {
				continue
			}

			padding := strings.Repeat("  ", n.Depth)
			nColor := cNode
			prefix := ""
			suffix := ""

			if n.Depth == 1 {
				subtree, _ := tree.Subtree(n.ID)
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

			util.Printfln(os.Stdout, "%s%s%s%s", padding, prefix, nColor(n.ID), suffix)
			if len(n.Vulnerabilities) > 0 {
				tw := new(tabwriter.Writer)
				tw.Init(os.Stdout, 10, 0, 2, ' ', 0)

				for i, v := range n.Vulnerabilities {
					var vColor func(...interface{}) string
					switch {
					case v.CVSSv2Score >= 9.0 || v.CVSSv3Score >= 9.0:
						vColor = cCritical
					case v.CVSSv2Score >= 7.0 || v.CVSSv3Score >= 7.0:
						vColor = cHigh
					case v.CVSSv2Score >= 4.0 || v.CVSSv3Score >= 4.0:
						vColor = cMedium
					default:
						vColor = cLow
					}

					util.Printfln(tw, "%s   %d\t%s\t%s\t%.1f/%.1f\t%s", padding, i+1,
						vColor(v.ID), vColor(v.Severity), v.CVSSv2Score, v.CVSSv3Score, v.CVEUrl)
				}
				tw.Flush()
			}
		}

		return nil
	},
}

func init() {
	vulTreeCmd.Flags().Bool("no-progress", false, "suppress in-progress output")
}
