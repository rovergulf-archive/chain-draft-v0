package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
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
			auth := getPassPhrase("Enter passphrase to encrypt new wallet:", true)

			w, err := wallets.InitWallets(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer w.Shutdown()

			wallet, err := w.AddWallet(auth)
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
			w, err := wallets.InitWallets(opts)
			if err != nil {
				return err
			}
			defer w.Shutdown()

			addresses, err := w.GetAllAddresses()
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

			auth := getPassPhrase("Enter password to decrypt the wallet:", false)

			w, err := wallets.InitWallets(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer w.Shutdown()

			wallet, err := w.GetWallet(common.HexToAddress(address))
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

func getPassPhrase(phrase string, confirmation bool) string {
	password, err := prompt.Stdin.PromptPassword(phrase)
	if err != nil {
		utils.Fatalf("Failed to read password: %v", err)
	}

	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			utils.Fatalf("Failed to read password confirmation: %v", err)
		}

		if password != confirm {
			utils.Fatalf("Passwords do not match")
		}
	}

	return password
}
