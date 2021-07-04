package commands

import (
	"context"
	"fmt"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/params"
	"github.com/rovergulf/rbn/pkg/resutil"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"path"
)

func getBackupDirPath() string {
	return path.Join(viper.GetString("data_dir"), viper.GetString("backup_dir"))
}

func getChainDbFilePath() string {
	return path.Join(viper.GetString("data_dir"), core.DbFileName)
}

func getWalletsDbFilePath() string {
	return path.Join(viper.GetString("data_dir"), wallets.DbWalletFile)
}

func getNodeDbFilePath() string {
	return path.Join(viper.GetString("data_dir"), node.DbFileName)
}

func prepareBlockchain(cmd *cobra.Command, args []string) error {
	bc, err := core.NewBlockchain(getBlockchainConfig(cmd))
	if err != nil {
		return err
	} else {
		blockChain = bc
	}
	fmt.Println(blockChain != nil)
	return nil
}

func prepareWalletsManager(cmd *cobra.Command, args []string) error {
	wm, err := wallets.NewManager(getBlockchainConfig(cmd))
	if err != nil {
		return err
	}

	accountManager = wm
	return nil
}

func prepareNode(cmd *cobra.Command, args []string) error {
	n, err := node.New(getBlockchainConfig(cmd))
	if err != nil {
		return err
	}

	if err := n.Init(context.Background()); err != nil {
		return err
	}

	localNode = n
	return nil
}

func getBlockchainConfig(cmd *cobra.Command) params.Options {
	opts := params.Options{
		Logger: logger,
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

func addNetworkIdFlag(cmd *cobra.Command) {
	cmd.Flags().String("network-id", params.MainNetworkId, "Chain network id")
	bindViperFlag(cmd, "network-id", "network-id")
}

func addAddressFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("address", "a", "", "Specify wallet address")
	cmd.MarkFlagRequired("address")
	bindViperFlag(cmd, "address", "address")
}
