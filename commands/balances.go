package commands

import (
	"fmt"
	"github.com/ethereum/go-ethereum/common"
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
		Use:     "list",
		Short:   "Lists all balances.",
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			defer blockChain.Shutdown()

			balances, err := blockChain.ListBalances()
			if err != nil {
				return err
			}

			return writeOutput(cmd, balances)
		},
	}

	addOutputFormatFlag(balancesListCmd)

	return balancesListCmd
}

// balancesGetCmd represents the balances get command
func balancesGetCmd() *cobra.Command {
	var balancesGetCmd = &cobra.Command{
		Use:     "get",
		Short:   "Get blockchain address balance.",
		PreRunE: prepareBlockchain,
		RunE: func(cmd *cobra.Command, args []string) error {
			address, _ := cmd.Flags().GetString("address")
			if len(address) > 0 {
				if !common.IsHexAddress(address) {
					return fmt.Errorf("invalid address")
				}
			}

			defer blockChain.Shutdown()

			balance, err := blockChain.GetBalance(common.HexToAddress(address))
			if err != nil {
				return err
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
