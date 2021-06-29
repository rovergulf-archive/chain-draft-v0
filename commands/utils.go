package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/rovergulf/rbn/pkg/resutil"
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

func getChainDbFilePath() string {
	return path.Join(getNodeDataPath(), core.DbFileName)
}

func getWalletsDbFilePath() string {
	return path.Join(getNodeDataPath(), wallets.DbWalletFile)
}

func getNodeDbFilePath() string {
	return path.Join(getNodeDataPath(), node.DbFileName)
}

func getBlockchainConfig(cmd *cobra.Command) config.Options {
	address := viper.GetString("address")
	nodeId := viper.GetString("node_id")
	minerAuth := viper.GetString("miner_auth")

	if len(nodeId) == 0 {
		nodeId, _ = cmd.Flags().GetString("node-id")
		viper.Set("node_id", nodeId)
	}

	if len(address) == 0 {
		address, _ = cmd.Flags().GetString("address")
		viper.Set("address", address)
	}

	if len(minerAuth) == 0 {
		minerAuth, _ = cmd.Flags().GetString("auth")
		viper.Set("miner_auth", minerAuth)
	}

	opts := config.Options{
		Address: address,
		NodeId:  nodeId,
		Logger:  logger,
	}

	opts.DbFilePath = getChainDbFilePath()
	opts.WalletsFilePath = getWalletsDbFilePath()
	opts.NodeFilePath = getNodeDbFilePath()

	return opts
}

func writeOutput(cmd *cobra.Command, v interface{}) error {
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "json" {
		return resutil.WriteJSON(os.Stdout, logger, v)
	} else {
		return resutil.WriteYAML(os.Stdout, logger, v)
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
