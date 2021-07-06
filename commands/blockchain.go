package commands

import (
	"context"
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
	blockchainCmd.AddCommand(blockchainGenesisCmd())

	return blockchainCmd
}

// initBlockchainCmd represents the run command
func initBlockchainCmd() *cobra.Command {
	var initBlockchainCmd = &cobra.Command{
		Use:     "init",
		Short:   "Start chain genesis block",
		Long:    ``,
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			defer blockChain.Shutdown()

			if err := blockChain.NewGenesisBlockWithRewrite(ctx); err != nil {
				return err
			}

			logger.Infow("Blockchain initialized.", "tip", blockChain.LastHash)

			return nil
		},
		TraverseChildren: true,
	}

	initBlockchainCmd.Flags().StringP("genesis", "g", "", "Genesis file path")
	initBlockchainCmd.MarkFlagRequired("genesis")
	bindViperFlag(initBlockchainCmd, "genesis", "genesis")

	return initBlockchainCmd
}

// blockchainListCmd represents the blocks list command
func blockchainListCmd() *cobra.Command {
	var blockchainListCmd = &cobra.Command{
		Use:     "list",
		Short:   "Lists all blocks.",
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer blockChain.Shutdown()
			maxLimit, _ := cmd.Flags().GetInt("limit")

			bci := blockChain.Iterator()

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

				if err := writeOutput(cmd, map[string]interface{}{
					"hash":      block.Hash,
					"prev_hash": block.PrevHash,
					"txs":       block.Transactions,
				}); err != nil {
					logger.Errorf("Unable to write block response: %s", err)
				}

				if len(block.PrevHash.Bytes()) == 0 {
					break
				}
			}

			return nil
		},
		TraverseChildren: true,
	}

	blockchainListCmd.Flags().Int("limit", 10, "Limit to show")
	addOutputFormatFlag(blockchainListCmd)

	return blockchainListCmd
}

// blockchainLastBlockCmd represents the blockchain last-block command
func blockchainLastBlockCmd() *cobra.Command {
	var blockchainLastBlockCmd = &cobra.Command{
		Use:     "last-block",
		Short:   "Show last block in the chain",
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer blockChain.Shutdown()

			block, err := blockChain.GetBlock(blockChain.LastHash)
			if err != nil {
				return err
			}

			return writeOutput(cmd, block)
		},
		TraverseChildren: true,
	}

	addOutputFormatFlag(blockchainLastBlockCmd)

	return blockchainLastBlockCmd
}

func blockchainGenesisCmd() *cobra.Command {
	var blockchainGenesisCmd = &cobra.Command{
		Use:     "show-genesis",
		Aliases: []string{"get-gen", "genesis"},
		Short:   "Show chain genesis",
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			defer blockChain.Shutdown()

			gen, err := blockChain.GetGenesis(ctx)
			if err != nil {
				return err
			}

			return writeOutput(cmd, gen)
		},
		TraverseChildren: true,
	}

	return blockchainGenesisCmd
}
