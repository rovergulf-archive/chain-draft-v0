package commands

import (
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

			bc, err := core.NewBlockchain(core.Options{
				DbFilePath: getDbFilePath(),
				Logger:     logger,
			})
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Db.Close()
			bci := bc.Iterator()

			for {
				block, err := bci.Next()
				if err != nil {
					return err
				}

				pow := core.NewProofOfWork(block)
				logger.Infof("Current block hash: '%x'; Previous hash: '%x'; Validated: %v",
					block.Hash, block.PrevHash, pow.Validate(),
				)

				if len(block.PrevHash) == 0 {
					break
				}

				return nil
			}

			return nil
		},
	}

	return blocksListCmd
}

// blocksAddCmd represents the blocks add command
func blocksAddCmd() *cobra.Command {
	var blocksAddCmd = &cobra.Command{
		Use:   "add",
		Short: "Add new blockchain block",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, _ := cmd.Flags().GetString("data")

			bc, err := core.NewBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Db.Close()

			logger.Infof("Add block with data: %s", data)

			// return bc.MineBlock(data)
			return nil
		},
	}

	blocksAddCmd.Flags().StringP("data", "d", "", "Specify block data")
	blocksAddCmd.MarkFlagRequired("data")

	return blocksAddCmd
}
