package commands

import (
	"fmt"
	"github.com/rovergulf/rbn/accounts"
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

			balance := 0
			UTXOs, err := bc.FindUTXO()
			if err != nil {
				return err
			}

			for _, out := range UTXOs {
				for i := range out.Outputs {
					logger.Infof("out.Outputs[%d].Value: %d", i, out.Outputs[i].Value)
				}
				balance += out.Outputs[0].Value
			}

			result := map[string]interface{}{
				"balance": balance,
			}
			return writeOutput(cmd, result)
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
			address := viper.GetString("address")
			if len(address) > 0 {
				if !accounts.ValidateAddress(address) {
					return fmt.Errorf("invalid address")
				}
			}

			bc, err := core.ContinueBlockchain(getBlockchainConfig(cmd))
			if err != nil {
				logger.Error("Unable to start blockchain: %s", err)
				return err
			}
			defer bc.Shutdown()

			balance := 0
			UTXOs, err := bc.FindUTXO()
			if err != nil {
				return err
			}

			for _, out := range UTXOs {
				fmt.Println("out.Outputs[0].Value:", out.Outputs[0].Value)
				balance += out.Outputs[0].Value
			}

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
