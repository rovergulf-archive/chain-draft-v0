package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
func runCmd() *cobra.Command {
	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run RNT node server",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			initBlockchain, _ := cmd.Flags().GetBool("init")

			var bc *core.Blockchain
			var err error

			if initBlockchain {
				bc, err = core.CreateBlockchain(getBlockchainConfig(cmd))
			} else {
				bc, err = core.NewBlockchain(getBlockchainConfig(cmd))
			}
			if err != nil {
				logger.Warnf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Db.Close()

			logger.Infow("Blockchain initialized.", "tip", bc.Tip)

			return nil
		},
	}

	addAddressFlag(runCmd)
	runCmd.Flags().Bool("init", false, "Init blockchain address")

	return runCmd
}

func init() {
	rootCmd.AddCommand(runCmd())
}
