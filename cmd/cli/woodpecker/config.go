package main

import (
	"fmt"
	"github.com/JackKCWong/go-woodpecker/cmd/cli/config"
	"github.com/JackKCWong/go-woodpecker/internal/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"os"
	"path"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Print current config file. If not found, prints a template",
	RunE: func(cmd *cobra.Command, args []string) error {
		userhome := os.Getenv("HOME")
		fp := path.Join(userhome, ".woodpecker.yaml")

		if err := viper.ReadInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error
				cfg, err := config.YamlTemplate.Open("template.yaml")
				if err != nil {
					return fmt.Errorf("failed to load config template: %w", err)
				}

				confYaml, err := io.ReadAll(cfg)
				if err != nil {
					return fmt.Errorf("failed to read config template: %w", err)
				}

				//util.Printfln(os.Stderr, "config not found. put the following template to %s/.woodpecker.yaml", os.Getenv("HOME"))
				util.Printfln(os.Stdout, string(confYaml))
				util.Printfln(os.Stdout, "the above template is written to [%s], fill in the blanks.", fp)

				err = os.WriteFile(fp, confYaml, 0700)
				if err != nil {
					return fmt.Errorf("failed to init config file [%s]: %w", fp, err)
				}

				return nil
			} else {
				// Config file was found but another error was produced
				return fmt.Errorf("failed to read config file: %w", err)
			}
		}

		cfg, err := util.ReadFile(fp)
		if err != nil {
			return fmt.Errorf("failed to read config file [%s]: %w", fp, err)
		}

		util.Printfln(os.Stdout, cfg)
		return nil
	},
}
