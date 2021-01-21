package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bigkevmcd/env-history/pkg/scanning"
	"github.com/go-git/go-git/v5"
)

func makeScanRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan-repo",
		Short: "scan-repo",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: accept list of strings for environments.

			r, err := git.PlainOpen(viper.GetString("repo-path"))
			if err != nil {
				log.Fatalf("failed to open the repository in %q: %s", viper.GetString("repo-path"), err)
			}
			result, err := scanning.Scan(
				r,
				viper.GetString("config-root"),
				viper.GetStringSlice("environments"))
			if err != nil {
				log.Fatal(err)
			}
			for k, v := range result {
				log.Printf("environment %q = %s", k, v)
			}
		},
	}
	cmd.Flags().String(
		"repo-path",
		"",
		"Path to git repository to interrogate",
	)
	logIfError(cmd.MarkFlagRequired("repo-path"))
	logIfError(viper.BindPFlag("repo-path", cmd.Flags().Lookup("repo-path")))

	cmd.Flags().StringSlice(
		"environments",
		nil,
		"Names of environments to find commits for",
	)
	logIfError(cmd.MarkFlagRequired("environments"))
	logIfError(viper.BindPFlag("environments", cmd.Flags().Lookup("environments")))

	cmd.Flags().String(
		"config-root",
		"",
		"path to start searching for commits from",
	)
	logIfError(viper.BindPFlag("config-root", cmd.Flags().Lookup("config-root")))
	return cmd
}
