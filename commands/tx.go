package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"time"
)

func init() {
	rootCmd.AddCommand(txCmd())
}

func txCmd() *cobra.Command {
	var txCmd = &cobra.Command{
		Use:              "tx",
		Short:            "Make send transaction or re-index current transactions",
		Long:             ``,
		SilenceUsage:     true,
		TraverseChildren: true,
	}

	txCmd.AddCommand(txReindexTxCmd())
	txCmd.AddCommand(txSendCmd())
	txCmd.AddCommand(txGetCmd())

	return txCmd
}

// txReindexTxCmd represents the reindex command
func txReindexTxCmd() *cobra.Command {
	var txReindexTxCmd = &cobra.Command{
		Use:   "reindex",
		Short: "Re-index transactions",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Warnf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			var count int
			logger.Infof("Done! There are %d transactions in the UTXO set.", count)

			return nil
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(txReindexTxCmd)

	return txReindexTxCmd
}

// txGetCmd represents the reindex command
func txGetCmd() *cobra.Command {
	var txGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Describe transaction by id",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Warnf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			tx, err := bc.FindTransaction([]byte(id))
			if err != nil {
				return err
			}

			return writeOutput(cmd, tx)
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(txGetCmd)
	addNodeIdFlag(txGetCmd)
	txGetCmd.Flags().String("id", "", "Transaction id")
	txGetCmd.MarkFlagRequired("id")

	return txGetCmd
}

// txSendCmd represents the send command
func txSendCmd() *cobra.Command {
	var txSendCmd = &cobra.Command{
		Use:   "send",
		Short: "Send coins to specified address",
		Args: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			amount, _ := cmd.Flags().GetInt("amount")
			fmt.Println(from, to, amount)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			mineNow, _ := cmd.Flags().GetBool("miner")
			from, _ := cmd.Flags().GetString("address")
			to, _ := cmd.Flags().GetString("to")
			amount, _ := cmd.Flags().GetInt64("amount")

			if !common.IsHexAddress(to) {
				return fmt.Errorf("recipient address is not Valid")
			}

			if !common.IsHexAddress(from) {
				return fmt.Errorf("sender address is not Valid")
			}

			if amount <= 0 {
				return fmt.Errorf("amount must be more than 0")
			}

			opts := getBlockchainConfig(cmd)

			bc, err := core.ContinueBlockchain(opts)
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			wm, err := wallets.NewManager(opts)
			if err != nil {
				return err
			}
			defer wm.Shutdown()

			wallet, err := wm.GetWallet(common.HexToAddress(from))
			if err != nil {
				return err
			}

			tx, err := core.NewTransaction(common.HexToAddress(from), common.HexToAddress(to), amount, 0, nil)
			if err != nil {
				return err
			}

			signedTx, err := wallet.SignTx(tx)
			if err != nil {
				return err
			}

			if mineNow {
				logger.Debug("Mine block")

				txs := []*core.SignedTx{signedTx}
				now := time.Now()

				b := core.NewBlock(bc.LastHash, bc.ChainLength.Uint64(), 0, now.Unix(), common.HexToAddress(from), txs)
				if err := b.SetHash(); err != nil {
					return err
				}

				block, err := bc.MineBlock(txs)
				if err != nil {
					return err
				}

				fmt.Println(b, block)
				return nil
			} else {
				return fmt.Errorf("not implemented")
				//return node.SendTx(viper.GetString("node_id"), tx)
			}
		},
		TraverseChildren: true,
	}

	txSendCmd.Flags().Bool("miner", false, "Mine block")
	txSendCmd.Flags().String("to", "", "Receiver address")
	txSendCmd.Flags().Int("amount", 0, "Transaction coin amount")
	txSendCmd.MarkFlagRequired("to")
	txSendCmd.MarkFlagRequired("amount")

	addAddressFlag(txSendCmd)
	addNodeIdFlag(txSendCmd)

	return txSendCmd
}
