package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/rovergulf/rbn/pkg/response"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

func getNodeDataPath() string {
	return path.Join(viper.GetString("data_dir"), viper.GetString("node_id"))
}

func getBackupDirPath() string {
	return path.Join(getNodeDataPath(), viper.GetString("backup_dir"))
}

func getDbFilePath() string {
	return path.Join(getNodeDataPath(), core.DbFileName)
}

func getWalletFilePath() string {
	return path.Join(getNodeDataPath(), wallets.DbWalletFile)
}

func getBlockchainConfig(cmd *cobra.Command) config.Options {
	address := viper.GetString("address")
	nodeId := viper.GetString("node_id")

	if len(nodeId) == 0 {
		nodeId, _ = cmd.Flags().GetString("node-id")
		viper.Set("node_id", nodeId)
	}

	if len(address) == 0 {
		address, _ = cmd.Flags().GetString("address")
		viper.Set("address", address)
	}

	opts := config.Options{
		Address: address,
		NodeId:  nodeId,
		Logger:  logger,
	}

	opts.DbFilePath = getDbFilePath()
	opts.WalletsFilePath = getWalletFilePath()

	return opts
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
	cmd.MarkFlagRequired("address")
	bindViperFlag(cmd, "address", "address")
}

func addNodeIdFlag(cmd *cobra.Command) {
	cmd.Flags().String("node-id", os.Getenv("NODE_ID"), "Blockchain node id")
	cmd.MarkFlagRequired("node-id")
	bindViperFlag(cmd, "node_id", "node-id")
}
