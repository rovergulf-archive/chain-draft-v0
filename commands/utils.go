package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/response"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

func getDbFilePath() string {
	return path.Join(viper.GetString("data_dir"), "chain.db")
}

func getBlockchainConfig(cmd *cobra.Command) core.Options {
	bcAddr, _ := cmd.Flags().GetString("address")

	return core.Options{
		DbFilePath: getDbFilePath(),
		Logger:     logger,
		Address:    bcAddr,
	}
}

func writeOutput(cmd *cobra.Command, v interface{}) error {
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "json" {
		return response.WriteJSON(os.Stdout, logger, v)
	} else {
		return response.WriteYAML(os.Stdout, logger, v)
	}
}

func bindViperFlag(cmd *cobra.Command, viperVal, flagName string) {
	if err := viper.BindPFlag(viperVal, cmd.Flags().Lookup(flagName)); err != nil {
		log.Printf("Failed to bind viper flag: %s", err)
	}
}

func bindViperPersistentFlag(cmd *cobra.Command, viperVal, flagName string) {
	if err := viper.BindPFlag(viperVal, cmd.PersistentFlags().Lookup(flagName)); err != nil {
		log.Printf("Failed to bind viper flag: %s", err)
	}
}

func addOutputFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("output", "o", "yaml", "specify output format (yaml/json)")
}

func addAddressFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "Blockchain address")
}
