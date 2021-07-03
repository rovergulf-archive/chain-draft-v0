package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
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

			useMnemonic, _ := cmd.Flags().GetBool("mnemonic")

			var auth string

			if !useMnemonic {
				input, err := getPassPhrase("Enter secret passphrase to encrypt the wallet:", false)
				if err != nil {
					return err
				}

				if len(input) < 6 {
					return fmt.Errorf("too weak, min 6 symbols length")
				}

				auth = input
			} else {
				// generate a random Mnemonic in English with 256 bits of entropy
				entropy, _ := bip39.NewEntropy(256)
				auth, _ = bip39.NewMnemonic(entropy)

				logger.Infof("Random Mnemonic passphrase to unlock wallet: \n\n\t%s\n", auth)
				logger.Warn("Save this passphrase to access your wallet.",
					"There is no way to recover it, but you can change it")
			}

			key, err := wallets.NewRandomKey()
			if err != nil {
				return err
			}

			wallet, err := wm.AddWallet(key, auth)
			if err != nil {
				return err
			}

			logger.Infof("Done! Wallet address: \n\n\t%s\n", wallet.Address())
			return nil
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(walletsNewCmd)
	bindViperFlag(walletsNewCmd, "node_id", "node-id")
	walletsNewCmd.MarkFlagRequired("node-id")

	walletsNewCmd.Flags().Bool("mnemonic", true, "Use mnemonic passphrase for wallet encrypting")

	return walletsNewCmd
}

func walletsUpdateAuthCmd() *cobra.Command {
	var walletsNewCmd = &cobra.Command{
		Use:   "update",
		Short: "Change wallet passphrase",
		RunE: func(cmd *cobra.Command, args []string) error {
			wm, err := wallets.NewManager(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			flagAddr, _ := cmd.Flags().GetString("address")
			if !common.IsHexAddress(flagAddr) {
				return fmt.Errorf("invalid address: %s", flagAddr)
			}
			addr := common.HexToAddress(flagAddr)

			useMnemonic, _ := cmd.Flags().GetBool("mnemonic")

			var newAuth string
			auth, err := getPassPhrase("Enter passphrase do decrypt wallet:", true)
			if err != nil {
				return err
			}

			if !useMnemonic {
				input, err := getPassPhrase("Enter old password:", false)
				if err != nil {
					return err
				}

				if len(input) < 6 {
					return fmt.Errorf("too weak, min 6 symbols length")
				}

				newAuth = input
			} else {
				// generate a random Mnemonic in English with 256 bits of entropy
				mnemonic, err := wallets.NewRandomMnemonic()
				if err != nil {
					return err
				}
				newAuth = mnemonic

				logger.Infof("Random Mnemonic passphrase to unlock wallet: \n\n\t%s\n", auth)
				logger.Warn("Save this passphrase to access your wallet.",
					"There is no way to recover it, but you can change it")
			}

			w, err := wm.GetWallet(addr, auth)
			if err != nil {
				logger.Errorf("Unable to get wallet: %s", err)
				return err
			}

			if _, err := wm.AddWallet(w.GetKey(), newAuth); err != nil {
				return err
			}

			logger.Infof("Done! Passphrase for account '%s' has changed!", addr.Hex())
			return nil
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(walletsNewCmd)
	addAddressFlag(walletsNewCmd)
	bindViperFlag(walletsNewCmd, "node_id", "node-id")
	walletsNewCmd.MarkFlagRequired("node-id")

	walletsNewCmd.Flags().Bool("mnemonic", true, "Use mnemonic passphrase for wallet encrypting")

	return walletsNewCmd
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

			auth, err := getPassPhrase("Enter passphrase do decrypt wallet:", true)
			if err != nil {
				return err
			}

			wm, err := wallets.NewManager(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			wallet, err := wm.GetWallet(common.HexToAddress(address), auth)
			if err != nil {
				logger.Errorf("Unable to get wallet: %s", err)
				return err
			}

			return writeOutput(cmd, wallet.GetKey())
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(walletsPrintPrivKeyCmd)
	addAddressFlag(walletsPrintPrivKeyCmd)
	addNodeIdFlag(walletsPrintPrivKeyCmd)

	return walletsPrintPrivKeyCmd
}

func walletsListCmd() *cobra.Command {
	var walletsListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists available wallet addresses.",
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

func getPassPhrase(message string, confirmation bool) (string, error) {
	auth, err := prompt.Stdin.PromptPassword(message)
	if err != nil {
		return "", err
	}

	if confirmation {
		confirm, err := prompt.Stdin.PromptPassword("Repeat password: ")
		if err != nil {
			return "", fmt.Errorf("failed to read passphrase confirmation: %v", err)
		}

		if auth != confirm {
			return "", fmt.Errorf("passphrases do not match")
		}
	}

	return auth, nil
}
