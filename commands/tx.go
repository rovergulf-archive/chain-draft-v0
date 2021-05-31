package commands

import (
	"fmt"
	"github.com/rovergulf/rbn/accounts"
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
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
	//txCmd.AddCommand(txGetCmd())

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

			UTXOSet := core.UTXOSet{Blockchain: bc}
			if err := UTXOSet.Reindex(); err != nil {
				return err
			}

			count, err := UTXOSet.CountTransactions()
			if err != nil {
				return err
			}

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

			//bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			//if err != nil {
			//	logger.Warnf("Unable to start blockchain: %s", err)
			//	return err
			//}
			//defer bc.Shutdown()

			logger.Infof("Not implemented: Get tx: %s", id)

			return nil
		},
		TraverseChildren: true,
	}

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
			amount, _ := cmd.Flags().GetInt("amount")

			if !accounts.ValidateAddress(to) {
				return fmt.Errorf("recipient address is not Valid")
			}

			if !accounts.ValidateAddress(from) {
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

			UTXOSet := core.UTXOSet{Blockchain: bc}

			wallets, err := accounts.InitWallets(opts)
			if err != nil {
				return err
			}
			defer wallets.Shutdown()

			wallet, err := wallets.GetWallet(from)
			if err != nil {
				return err
			}

			tx, err := core.NewTransaction(wallet, to, amount, &UTXOSet)
			if err != nil {
				return err
			}

			if mineNow {
				logger.Debug("Mine block")
				cbTx := core.CoinbaseTx(from, "")
				txs := []*core.Transaction{cbTx, tx}

				block, err := bc.MineBlock(txs)
				if err != nil {
					return err
				}

				return UTXOSet.Update(block)
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
