package commands

import (
	"github.com/spf13/cobra"
)

// nodeCmd represents the node command
var nodeCmd = &cobra.Command{
	Use:          "node",
	Short:        "Node maintenance",
	Long:         ``,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(nodeCmd)
}

func nodeStatusCmd() *cobra.Command {
	var nodeStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Get node status",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			logger.Info("node status called")
		},
	}

	return nodeStatusCmd
}
