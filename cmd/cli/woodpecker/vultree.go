package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var vulTreeCmd = &cobra.Command{
	Use:     "tree",
	Short:   "Print dependency tree with CVEs",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		updater := maven.NewUpdater("pom.xml", maven.UpdaterOpts{
			Verbose: verbose,
		})

		tree, err := updater.DependencyTree()
		if err != nil {
			return err
		}

		for _, n := range tree.Nodes {
			padding := strings.Repeat("  ", n.Depth)
			nColor := color.WhiteString
			if n.Scope == "test" {
				continue
			}

			prefix := ""
			switch {
			case n.Depth == 1:
				prefix = "+ "
			case n.Depth > 1:
				prefix = "- "
			}

			util.Printfln(os.Stdout, "%s%s%s", padding, prefix, nColor(n.ID))
			if len(n.Vulnerabilities) > 0 {
				for _, v := range n.Vulnerabilities {
					var vColor func(string, ...interface{}) string
					switch {
					case v.CVSSv2Score >= 9.0 || v.CVSSv3Score >= 9.0:
						vColor = color.HiRedString
					case v.CVSSv2Score >= 7.0 || v.CVSSv3Score >= 7.0:
						vColor = color.RedString
					case v.CVSSv2Score >= 4.0 || v.CVSSv3Score >= 4.0:
						vColor = color.YellowString
					default:
						vColor = color.BlueString
					}

					util.Printfln(os.Stdout, "%s  * %s\t%s", padding, vColor(v.ID), vColor(v.Severity))
				}
			}
		}

		return nil
	},
}

func init() {
	vulTreeCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
