package commands

import (
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
)

// balancesCmd represents the balances command
var balancesCmd = &cobra.Command{
	Use:          "balances",
	Short:        "A brief description of your command",
	Long:         ``,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(balancesCmd)
	balancesCmd.AddCommand(balancesListCmd())
	balancesCmd.AddCommand(balancesGetCmd())
}

// balancesListCmd represents the balances list command
func balancesListCmd() *cobra.Command {
	var balancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "Lists all balances.",
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.NewBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Db.Close()

			logger.Info("balances list called")
			return nil
		},
	}

	addOutputFormatFlag(balancesListCmd)

	return balancesListCmd
}

// balancesGetCmd represents the balances get command
func balancesGetCmd() *cobra.Command {
	var balancesGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get blockchain address balance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			bc, err := core.NewBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Db.Close()

			address, _ := cmd.Flags().GetString("address")

			balance := 0
			UTXOs := bc.FindUTXO(address)

			for _, out := range UTXOs {
				balance += out.Value
			}

			return writeOutput(cmd, map[string]interface{}{
				"address": address,
				"balance": balance,
			})
		},
	}

	addAddressFlag(balancesGetCmd)
	addOutputFormatFlag(balancesGetCmd)

	return balancesGetCmd
}
