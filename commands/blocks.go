package commands

import (
	"fmt"
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
)

// blocksCmd represents the blocks command
var blocksCmd = &cobra.Command{
	Use:          "blocks",
	Short:        "A brief description of your command",
	Long:         ``,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(blocksCmd)
	blocksCmd.AddCommand(blocksListCmd())
	blocksCmd.AddCommand(blocksAddCmd())
}

// blocksListCmd represents the blocks list command
func blocksListCmd() *cobra.Command {
	var blocksListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all blocks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("blocks list called")

			bc, err := core.ContinueBlockchain(core.Options{
				DbFilePath: getDbFilePath(),
				Logger:     logger,
			})
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()
			bci := bc.Iterator()

			for {
				block, err := bci.Next()
				if err != nil {
					return err
				}

				pow := core.NewProofOfWork(block)
				if err := writeOutput(cmd, map[string]interface{}{
					"hash":      fmt.Sprintf("%x", block.Hash),
					"prev_hash": fmt.Sprintf("%x", block.PrevHash),
					"pow":       fmt.Sprintf("%v", pow.Validate()),
				}); err != nil {
					logger.Errorf("Unable to write block response: %s", err)
				}

				if len(block.PrevHash) == 0 {
					break
				}
			}

			return nil
		},
	}

	addOutputFormatFlag(blocksListCmd)

	return blocksListCmd
}

// blocksAddCmd represents the blocks add command
func blocksAddCmd() *cobra.Command {
	var blocksAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add new blockchain block",
		RunE: func(cmd *cobra.Command, args []string) error {
			from, _ := cmd.Flags().GetString("from")
			to, _ := cmd.Flags().GetString("to")
			amount, _ := cmd.Flags().GetInt("amount")

			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			tx, err := core.NewTransaction(from, to, amount, bc)
			if err != nil {
				return err
			}

			newBlock, err := bc.MineBlock([]*core.Transaction{tx})
			if err != nil {
				return err
			} else {
				return writeOutput(cmd, newBlock)
			}
		},
	}

	blocksAddCmd.Flags().String("from", "", "Sender address")
	blocksAddCmd.Flags().String("to", "", "Receiver address")
	blocksAddCmd.Flags().String("amount", "", "Transaction coin amount")
	blocksAddCmd.MarkFlagRequired("from")
	blocksAddCmd.MarkFlagRequired("to")
	blocksAddCmd.MarkFlagRequired("amount")

	return blocksAddCmd
}
