package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/client"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/proto"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func txCmd() *cobra.Command {
	var txCmd = &cobra.Command{
		Use:              "tx",
		Short:            "Make send transaction or re-index current transactions",
		Long:             ``,
		SilenceUsage:     true,
		TraverseChildren: true,
	}

	txCmd.AddCommand(txSendCmd())
	txCmd.AddCommand(txGetCmd())

	return txCmd
}

// txGetCmd represents the reindex command
func txGetCmd() *cobra.Command {
	var txGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "Describe transaction by id",
		Long:    ``,
		PreRunE: prepareBlockChain,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer blockChain.Shutdown()
			id, err := cmd.Flags().GetString("id")
			if err != nil {
				return err
			}

			tx, err := blockChain.FindTransaction(common.HexToHash(id))
			if err != nil {
				return err
			}

			return writeOutput(cmd, tx)
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(txGetCmd)
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
			amount, _ := cmd.Flags().GetFloat64("amount")

			if !common.IsHexAddress(to) {
				return fmt.Errorf("recipient address is not Valid")
			}

			if !common.IsHexAddress(from) {
				return fmt.Errorf("sender address is not Valid")
			}

			if amount <= 0 {
				return fmt.Errorf("amount must be more than 0")
			}

			auth, err := prompt.Stdin.PromptPassword("Enter passphrase to decrypt wallet:")
			if err != nil {
				return err
			}

			c, err := client.NewClient(ctx, logger, viper.GetString("network.addr"))
			if err != nil {
				return err
			}
			defer c.Stop()

			addTxReq := node.TxAddRequest{
				From:    from,
				FromPwd: auth,
				To:      to,
				Value:   amount,
				Data:    nil,
			}

			callData, err := json.Marshal(addTxReq)
			if err != nil {
				return err
			}

			res, err := c.MakeCall(ctx, proto.Command_Add, proto.Entity_Transaction, callData)
			if err != nil {
				return err
			}

			return writeOutput(cmd, res)
		},
		TraverseChildren: true,
	}

	txSendCmd.Flags().String("to", "", "Receiver address")
	txSendCmd.Flags().Int("amount", 0, "Transaction coin amount")
	txSendCmd.MarkFlagRequired("to")
	txSendCmd.MarkFlagRequired("amount")

	addAddressFlag(txSendCmd)

	return txSendCmd
}
