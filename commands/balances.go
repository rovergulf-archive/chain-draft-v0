package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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
			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			return writeOutput(cmd, bc.Balances)
		},
	}

	addOutputFormatFlag(balancesListCmd)
	addNodeIdFlag(balancesListCmd)

	return balancesListCmd
}

// balancesGetCmd represents the balances get command
func balancesGetCmd() *cobra.Command {
	var balancesGetCmd = &cobra.Command{
		Use:   "get",
		Short: "Get blockchain address balance.",
		RunE: func(cmd *cobra.Command, args []string) error {
			address, _ := cmd.Flags().GetString("address")
			if len(address) > 0 {
				if !common.IsHexAddress(address) {
					return fmt.Errorf("invalid address")
				}
			}

			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Errorf("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			balanceSet := core.Balances{
				Blockchain: *bc,
			}

			balance, err := balanceSet.GetBalance(common.HexToAddress(address))
			if err != nil {
				return err
			}

			//UTXOs, err := UTXOSet.FindUnspentTransactions(pubKeyHash)
			//if err != nil {
			//	return err
			//}
			//
			//for _, out := range UTXOs {
			//	balance += out.Value
			//}

			return writeOutput(cmd, map[string]interface{}{
				"address": address,
				"balance": balance,
			})
		},
	}

	addAddressFlag(balancesGetCmd)
	addNodeIdFlag(balancesGetCmd)

	addOutputFormatFlag(balancesGetCmd)

	return balancesGetCmd
}
