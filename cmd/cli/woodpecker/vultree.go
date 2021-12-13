package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var (
	cCritical     = color.New(color.FgRed, color.BgHiYellow).SprintFunc()
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
		verbose, _ := cmd.Flags().GetBool("verbose")
		link, _ := cmd.Flags().GetBool("link")
		updater := maven.NewUpdater("pom.xml", maven.UpdaterOpts{
			Verbose: verbose,
		})

		tree, err := updater.DependencyTree()
		if err != nil {
			return err
		}

		for _, n := range tree.Nodes {
			if n.Scope == "test" {
				continue
			}

			padding := strings.Repeat("  ", n.Depth)
			nColor := cNode
			prefix := ""
			suffix := ""

			if n.ShouldUpdate {
				nColor = cShouldUpdate
				suffix = "\t\t\t<-----\tupdate this"
			}

			switch {
			case n.Depth == 1:
				prefix = "+ "
			case n.Depth > 1:
				prefix = "- "
			}

			util.Printfln(os.Stdout, "%s%s%s%s", padding, prefix, nColor(n.ID), suffix)
			if len(n.Vulnerabilities) > 0 {
				for _, v := range n.Vulnerabilities {
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

					util.Printfln(os.Stdout, "%s   * %s\t%s\t%s", padding,
						vColor(v.ID), vColor(v.Severity),
						map[bool]string{
							true:  v.CVEUrl,
							false: "",
						}[link])
				}
			}
		}

		return nil
	},
}

func init() {
	vulTreeCmd.Flags().BoolP("link", "l", false, "show ref links")
}
