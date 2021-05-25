package commands

import (
	"fmt"
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(blockchainCmd())
}

// blockchainCmd represents the run command
func blockchainCmd() *cobra.Command {
	var blockchainCmd = &cobra.Command{
		Use:          "blockchain",
		Short:        "Blockchain operations",
		Long:         ``,
		SilenceUsage: true,
	}

	blockchainCmd.AddCommand(initBlockchainCmd())
	blockchainCmd.AddCommand(blockchainListCmd())
	blockchainCmd.AddCommand(blockchainLastBlockCmd())

	return blockchainCmd
}

// initBlockchainCmd represents the run command
func initBlockchainCmd() *cobra.Command {
	var initBlockchainCmd = &cobra.Command{
		Use:   "init",
		Short: "Init blockchain",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.InitBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Warnf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			logger.Infow("Blockchain initialized.", "tip", bc.LastHash)

			return nil
		},
		TraverseChildren: true,
	}

	addAddressFlag(initBlockchainCmd)
	addNodeIdFlag(initBlockchainCmd)

	return initBlockchainCmd
}

// blockchainListCmd represents the blocks list command
func blockchainListCmd() *cobra.Command {
	var blockchainListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all blocks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			maxLimit, _ := cmd.Flags().GetInt("limit")

			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()
			bci := bc.Iterator()

			var limit int
			for {
				block, err := bci.Next()
				if err != nil {
					return err
				}

				if maxLimit == limit {
					break
				}
				limit++

				pow := core.NewProofOfWork(block)

				if err := writeOutput(cmd, map[string]interface{}{
					"hash":      block.GetHash(),
					"prev_hash": block.GetPrevHash(),
					"pow":       fmt.Sprintf("%v", pow.Validate()),
					"txs":       block.Transactions,
				}); err != nil {
					logger.Errorf("Unable to write block response: %s", err)
				}

				if len(block.PrevHash) == 0 {
					break
				}
			}

			return nil
		},
		TraverseChildren: true,
	}

	blockchainListCmd.Flags().Int("limit", 10, "Limit to show")
	addNodeIdFlag(blockchainListCmd)
	addOutputFormatFlag(blockchainListCmd)

	return blockchainListCmd
}

// blockchainLastBlockCmd represents the blockchain last-block command
func blockchainLastBlockCmd() *cobra.Command {
	var blockchainLastBlockCmd = &cobra.Command{
		Use:   "last-block",
		Short: "Show last block in the chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			block, err := bc.GetBlock(bc.LastHash)
			if err != nil {
				return err
			}

			return writeOutput(cmd, map[string]interface{}{
				"hash": block.GetHash(),
			})
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(blockchainLastBlockCmd)
	addOutputFormatFlag(blockchainLastBlockCmd)

	return blockchainLastBlockCmd
}
