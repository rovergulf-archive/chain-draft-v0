package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/rpc"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			//data, _ := cmd.Flags().GetBool("data")
			//dataFormat, _ := cmd.Flags().GetString("data-format")
			from, _ := cmd.Flags().GetString("address")
			to, _ := cmd.Flags().GetString("to")
			amount, _ := cmd.Flags().GetUint64("amount")

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

			auth, err := prompt.Stdin.PromptPassword("Enter old password:")
			if err != nil {
				return err
			}

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

			fromAddr := common.HexToAddress(from)

			wallet, err := wm.GetWallet(fromAddr, auth)
			if err != nil {
				logger.Errorf("Unable to get wallet: %s", err)
				return err
			}

			client, err := node.NewClient(ctx, logger, viper.GetString("network.addr"))
			if err != nil {
				return err
			}
			defer client.Stop()

			addTxReq := node.TxAddReq{
				From:    from,
				FromPwd: wallet.GetPassphrase(),
				To:      to,
				Value:   amount,
				Data:    nil,
			}

			callData, err := json.Marshal(addTxReq)
			if err != nil {
				return err
			}

			res, err := client.RpcCall(ctx, &rpc.CallRequest{
				Cmd:  rpc.CallRequest_TX_ADD,
				Data: callData,
			})
			if err != nil {
				return err
			}

			return writeOutput(cmd, res)
			//return node.SendTx(viper.GetString("node_id"), tx)
		},
		TraverseChildren: true,
	}

	txSendCmd.Flags().String("to", "", "Receiver address")
	txSendCmd.Flags().Int("amount", 0, "Transaction coin amount")
	txSendCmd.MarkFlagRequired("to")
	txSendCmd.MarkFlagRequired("amount")

	addAddressFlag(txSendCmd)
	addNodeIdFlag(txSendCmd)

	return txSendCmd
}
