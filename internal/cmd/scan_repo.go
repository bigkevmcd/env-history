package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/bigkevmcd/env-history/pkg/scanning"
)

func makeScanRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scan-repo --repo-path",
		Short: "scan-repo",
		Run: func(cmd *cobra.Command, args []string) {
			scanning.Scan(viper.GetString("repo-path"))
		},
	}
	cmd.Flags().String(
		"repo-path",
		"",
		"Path to git repository to interrogate",
	)
	logIfError(cmd.MarkFlagRequired("repo-path"))
	logIfError(viper.BindPFlag("repo-path", cmd.Flags().Lookup("repo-path")))
	cmd.Flags().String(
		"config-root",
		"",
		"path to start searching for commits from",
	)
	logIfError(viper.BindPFlag("config-root", cmd.Flags().Lookup("config-root")))
	return cmd
}
