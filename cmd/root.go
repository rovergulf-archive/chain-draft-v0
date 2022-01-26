package cmd

import (
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "rbn",
	Short:   "Rovergulf BlockChain API Interface",
	Long:    `Rovergulf BlockChain Network `,
	Version: "0.0.1-dev",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := cmd.Flags().GetBool("version")
		if ver {
			return writeOutput(cmd, cmd.Version)
		} else {
			return cmd.Usage()
		}
	},
	SilenceUsage: true,
}
