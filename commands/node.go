package commands

import (
	"github.com/rovergulf/rbn/node"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(nodeCmd())
}

// nodeCmd represents the node command
func nodeCmd() *cobra.Command {
	var nodeCmd = &cobra.Command{
		Use:              "node",
		Short:            "Node maintenance",
		Long:             ``,
		SilenceUsage:     true,
		TraverseChildren: true,
	}

	nodeCmd.AddCommand(nodeRunCmd())

	return nodeCmd
}

func nodeRunCmd() *cobra.Command {
	var nodeRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Rovergulf Blockchain Network peer node",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := node.New(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}

			return n.Run()
		},
		TraverseChildren: true,
	}

	nodeRunCmd.Flags().StringP("address", "a", "", "Blockchain address")
	addNodeIdFlag(nodeRunCmd)

	// node
	nodeRunCmd.Flags().Bool("miner", false, "Enable miner")

	nodeRunCmd.Flags().String("net-addr", "127.0.0.1", "Network discovery address")
	bindViperFlag(nodeRunCmd, "network.addr", "net-addr")
	nodeRunCmd.Flags().Int("net-port", 9420, "Network connection port")
	bindViperFlag(nodeRunCmd, "network.port", "net-port")

	nodeRunCmd.Flags().String("node-addr", "127.0.0.1", "Node address would listen to")
	bindViperFlag(nodeRunCmd, "node.addr", "node-addr")
	nodeRunCmd.Flags().Int("node-port", 9069, "Node port would listen to")
	bindViperFlag(nodeRunCmd, "node.port", "node-port")

	return nodeRunCmd
}
