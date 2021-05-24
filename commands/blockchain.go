package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
)

// initBlockchainCmd represents the run command
func initBlockchainCmd() *cobra.Command {
	var blockchainCmd = &cobra.Command{
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
	}

	addAddressFlag(blockchainCmd)
	blockchainCmd.Flags().Bool("init", false, "Init blockchain address")

	return blockchainCmd
}

// runCmd represents the run command
func runCmd() *cobra.Command {
	var blockchainCmd = &cobra.Command{
		Use:   "run",
		Short: "Continue existing blockchain",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Warnf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			logger.Infow("Blockchain initialized.", "last_hash", bc.LastHash)

			return nil
		},
	}

	addAddressFlag(blockchainCmd)

	return blockchainCmd
}

func init() {
	rootCmd.AddCommand(initBlockchainCmd())
	rootCmd.AddCommand(runCmd())
}
