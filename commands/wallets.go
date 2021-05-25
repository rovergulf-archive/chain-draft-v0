package commands

import (
	"fmt"
	"github.com/rovergulf/rbn/accounts"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletsCmd())
}

// walletsCmd represents the wallet command
func walletsCmd() *cobra.Command {
	var walletsCmd = &cobra.Command{
		Use:              "wallets",
		Short:            "Wallet related operations",
		Long:             ``,
		SilenceUsage:     true,
		TraverseChildren: true,
	}

	walletsCmd.AddCommand(walletsNewCmd())
	walletsCmd.AddCommand(walletsListCmd())
	walletsCmd.AddCommand(walletsPrintPrivKeyCmd())

	return walletsCmd
}

func walletsNewCmd() *cobra.Command {
	var walletsNewCmd = &cobra.Command{
		Use:   "new",
		Short: "Creates a new wallet.",
		RunE: func(cmd *cobra.Command, args []string) error {
			wallets, err := accounts.InitWallets(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer wallets.Shutdown()

			newWallet, err := wallets.AddWallet()
			if err != nil {
				return err
			}

			address, err := newWallet.Address()
			if err != nil {
				return err
			}

			logger.Infof("Done! Wallet address: \n\t%s", address)
			return nil
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(walletsNewCmd)
	bindViperFlag(walletsNewCmd, "node_id", "node-id")
	walletsNewCmd.MarkFlagRequired("node-id")

	return walletsNewCmd
}

func walletsListCmd() *cobra.Command {
	var walletsListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists available wallets.",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := getBlockchainConfig(cmd)
			wallets, err := accounts.InitWallets(opts)
			if err != nil {
				return err
			}
			defer wallets.Shutdown()

			addresses, err := wallets.GetAllAddresses()
			if err != nil {
				return err
			}

			return writeOutput(cmd, map[string]interface{}{
				"_node_id":  opts.NodeId,
				"addresses": addresses,
			})
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(walletsListCmd)
	addNodeIdFlag(walletsListCmd)

	return walletsListCmd
}

func walletsPrintPrivKeyCmd() *cobra.Command {
	var walletsPrintPrivKeyCmd = &cobra.Command{
		Use:   "print-pk",
		Short: "Unlocks keystore file and prints the Private + Public keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			address, _ := cmd.Flags().GetString("address")

			wallets, err := accounts.InitWallets(config.Options{
				Logger: logger,
			})
			if err != nil {
				return err
			}
			defer wallets.Shutdown()

			wallet, err := wallets.GetWallet(address)
			if err != nil {
				return err
			}

			addr, err := wallet.StringAddr()
			if err != nil {
				return err
			}

			return writeOutput(cmd, map[string]interface{}{
				"address":    addr,
				"pk_cure":    fmt.Sprintf("%x", wallet.PrivateKey.Curve.Params()),
				"public_key": fmt.Sprintf("%x", wallet.PublicKey),
			})
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(walletsPrintPrivKeyCmd)
	addAddressFlag(walletsPrintPrivKeyCmd)
	addNodeIdFlag(walletsPrintPrivKeyCmd)

	return walletsPrintPrivKeyCmd
}

func getPassPhrase(phrase string, confirmation bool) string {
	return ""
}
