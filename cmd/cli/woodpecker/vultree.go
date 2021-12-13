package main

import (
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
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
			prefix := strings.Repeat("  ", n.Depth)
			util.Printfln(os.Stdout, "%s%s", prefix, n.ID)
			if len(n.Vulnerabilities) > 0 {
				for _, v := range n.Vulnerabilities {
					util.Printfln(os.Stdout, "%s  - %s\t%s", prefix, v.ID, v.Severity)
				}
			}
		}

		return nil
	},
}

func init() {
	vulTreeCmd.Flags().BoolP("verbose", "v", false, "verbose output")
}
