package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/tyler-smith/go-bip39"
	"io/ioutil"
	"path"
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
	walletsCmd.AddCommand(walletsUpdateAuthCmd())
	walletsCmd.AddCommand(walletsListCmd())
	walletsCmd.AddCommand(walletsPrintPrivKeyCmd())
	//walletsCmd.AddCommand(walletsImportCmd())

	return walletsCmd
}

func walletsListCmd() *cobra.Command {
	var walletsListCmd = &cobra.Command{
		Use:     "list",
		Short:   "Lists available wallet addresses.",
		PreRunE: prepareWalletsManager,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := getBlockchainConfig(cmd)
			defer accountManager.Shutdown()

			addresses, err := accountManager.GetAllAddresses()
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

	return walletsListCmd
}

func walletsPrintPrivKeyCmd() *cobra.Command {
	var walletsPrintPrivKeyCmd = &cobra.Command{
		Use:   "print-pk",
		Short: "Unlocks keystore file and prints the Private + Public keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer accountManager.Shutdown()

			address, _ := cmd.Flags().GetString("address")
			if !common.IsHexAddress(address) {
				return fmt.Errorf("bad address format")
			}

			auth, err := getPassPhrase("Enter passphrase do decrypt wallet:", true)
			if err != nil {
				return err
			}

			wallet, err := accountManager.GetWallet(common.HexToAddress(address), auth)
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

	return walletsPrintPrivKeyCmd
}

func walletsNewCmd() *cobra.Command {
	var walletsNewCmd = &cobra.Command{
		Use:     "new",
		Short:   "Creates a new wallet.",
		PreRunE: prepareWalletsManager,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer accountManager.Shutdown()

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

			wallet, err := accountManager.AddWallet(key, auth)
			if err != nil {
				return err
			}

			logger.Infof("Done! Wallet address: \n\n\t%s\n", wallet.Address())
			return nil
		},
		TraverseChildren: true,
	}

	walletsNewCmd.Flags().Bool("mnemonic", true, "Use mnemonic passphrase for wallet encrypting")

	return walletsNewCmd
}

func walletsUpdateAuthCmd() *cobra.Command {
	var walletsNewCmd = &cobra.Command{
		Use:   "update",
		Short: "Change wallet passphrase",
		RunE: func(cmd *cobra.Command, args []string) error {
			defer accountManager.Shutdown()

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

			w, err := accountManager.GetWallet(addr, auth)
			if err != nil {
				logger.Errorf("Unable to get wallet: %s", err)
				return err
			}

			if _, err := accountManager.AddWallet(w.GetKey(), newAuth); err != nil {
				return err
			}

			logger.Infof("Done! Passphrase for account '%s' has changed!", addr.Hex())
			return nil
		},
		TraverseChildren: true,
	}

	addAddressFlag(walletsNewCmd)

	walletsNewCmd.Flags().Bool("mnemonic", true, "Use mnemonic passphrase for wallet encrypting")

	return walletsNewCmd
}

func walletsImportCmd() *cobra.Command {
	walletsRecoverCmd := &cobra.Command{
		Use:     "import",
		Short:   "Imports key from specified CryptoJSON file to keystore",
		Long:    ``,
		PreRunE: prepareWalletsManager,
		RunE: func(cmd *cobra.Command, args []string) error {
			//ctx, cancel := context.WithCancel(context.Background())
			//defer cancel()
			defer accountManager.Shutdown()

			auth, err := getPassPhrase("Enter passphrase do decrypt wallet:", false)
			if err != nil {
				return err
			}

			filePath, _ := cmd.Flags().GetString("file")
			if path.Ext(filePath) != ".json" {
				return fmt.Errorf("file extension must be json")
			}

			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				return err
			}

			key, err := keystore.DecryptKey(data, auth)
			if err != nil {
				return err
			}

			w, err := accountManager.AddWallet(key, auth)
			if err != nil {
				return err
			}

			logger.Info("Successfully imported '%s' account into keystore", w.Address())
			return nil
		},
		TraverseChildren: true,
	}

	walletsRecoverCmd.Flags().StringP("file", "f", "", "Specify key file path to decode")
	walletsRecoverCmd.MarkFlagRequired("file")

	return walletsRecoverCmd
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
