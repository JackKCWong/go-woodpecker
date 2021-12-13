package main

import (
	"context"
	"github.com/JackKCWong/go-woodpecker/internal/spi/maven"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var vulTreeCmd = &cobra.Command{
	Use:     "tree",
	Short:   "Print dependency tree with CVEs",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		mvn := maven.Mvn{
			POM: "pom.xml",
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		temp, err := os.CreateTemp(os.TempDir(), "dtree")
		if err != nil {
			return err
		}

		stdout, errors := mvn.DependencyTree(ctx, temp.Name())
		verbose, _ := cmd.Flags().GetBool("verbose")
		if verbose {
			go util.DrainLines(os.Stdout, stdout)
		}

		err = <-errors
		if err != nil {
			return err
		}

		return nil
	},
}
