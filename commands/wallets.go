package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
	"os"
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
			wm, err := wallets.NewManager(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			// generate a random Mnemonic in English with 256 bits of entropy
			entropy, _ := bip39.NewEntropy(256)
			mnemonic, _ := bip39.NewMnemonic(entropy)

			logger.Infof("Random Mnemonic passphrase to unlock wallet: \n\n\t%s\n", mnemonic)
			logger.Warn("Save this passphrase to access your wallet.",
				"There is no way to recover it, but you can change it")

			wallet, err := wm.AddWallet(mnemonic)
			if err != nil {
				return err
			}

			logger.Infof("Done! Wallet address: \n\t%s", wallet.Address)
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
			wm, err := wallets.NewManager(opts)
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			addresses, err := wm.GetAllAddresses()
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
			if !common.IsHexAddress(address) {
				return fmt.Errorf("bad address format")
			}

			auth, err := prompt.Stdin.PromptPassword("Enter password to decrypt the wallet:")
			if err != nil {
				return err
			}

			wm, err := wallets.NewManager(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			wallet, err := wm.GetWallet(common.HexToAddress(address))
			if err != nil {
				return err
			}

			key, err := keystore.DecryptKey(wallet.Data, auth)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			return writeOutput(cmd, map[string]interface{}{
				"address": wallet.Address,
				"key":     key,
			})
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(walletsPrintPrivKeyCmd)
	addAddressFlag(walletsPrintPrivKeyCmd)
	addNodeIdFlag(walletsPrintPrivKeyCmd)

	return walletsPrintPrivKeyCmd
}
