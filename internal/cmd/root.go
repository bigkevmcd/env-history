package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	cobra.OnInitialize(initConfig)
}

func logIfError(e error) {
	if e != nil {
		log.Printf("%+v\n", e)
	}
}

func makeRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "env-history",
	}
	cmd.AddCommand(makeScanRepoCmd())
	return cmd
}

func initConfig() {
	viper.AutomaticEnv()
}

// Execute is the main entry point into this component.
func Execute() {
	logIfError(makeRootCmd().Execute())
}
