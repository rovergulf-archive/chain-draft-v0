package commands

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/console/prompt"
	"github.com/rovergulf/rbn/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
	"strconv"
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
	nodeCmd.AddCommand(nodeStopCmd())
	nodeCmd.AddCommand(nodeAccountDumpCmd())

	return nodeCmd
}

func nodeRunCmd() *cobra.Command {
	var nodeRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Rovergulf Blockchain Network peer node",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			addr, _ := cmd.Flags().GetString("address")

			if len(addr) > 0 {
				auth, err := prompt.Stdin.PromptPassword("Enter passphrase do decrypt miner wallet:")
				if err != nil {
					return err
				}

				viper.Set("auth", auth)
			}

			n, err := node.New(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}

			if err := n.Init(); err != nil {
				return err
			}

			return n.Run(ctx)
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(nodeRunCmd)

	nodeRunCmd.Flags().StringP("address", "a", "", "Node account address")

	// node
	nodeRunCmd.Flags().String("miner", "", "Specify miner account")

	nodeRunCmd.Flags().String("net-addr", "127.0.0.1:9420", "Network discovery address")
	bindViperFlag(nodeRunCmd, "network.addr", "net-addr")

	nodeRunCmd.Flags().String("node-addr", "127.0.0.1", "Node address would listen to")
	bindViperFlag(nodeRunCmd, "node.addr", "node-addr")
	nodeRunCmd.Flags().Int("node-port", 9420, "Node port would listen to")
	bindViperFlag(nodeRunCmd, "node.port", "node-port")
	nodeRunCmd.Flags().String("http-addr", "127.0.0.1", "Node address would listen to")
	bindViperFlag(nodeRunCmd, "http.addr", "http-addr")
	nodeRunCmd.Flags().Int("http-port", 9469, "Node port would listen to")
	bindViperFlag(nodeRunCmd, "http.port", "http-port")

	return nodeRunCmd
}

func nodeStopCmd() *cobra.Command {
	var nodeStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops running node using saved process id",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			pidValue, err := ioutil.ReadFile(viper.GetString("pid_file"))
			if err != nil {
				if err == os.ErrNotExist {
					logger.Info("Daemon not started")
					return nil
				}
				return err
			}

			pid, err := strconv.Atoi(string(pidValue))
			if err != nil {
				return err
			}

			p, err := os.FindProcess(pid)
			if err != nil {
				return err
			}

			if err := p.Signal(os.Interrupt); err != nil {
				logger.Errorf("Unable to stop daemon: %s", err)
				return err
			} else {
				logger.Info("Successfully stopped daemon")
				return os.Remove(viper.GetString("pid_file"))
			}
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(nodeStopCmd)

	return nodeStopCmd
}

func nodeAccountDumpCmd() *cobra.Command {
	nodeAccountDumpCmd := &cobra.Command{
		Use:   "account-dump",
		Short: "Export account key",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			n, err := node.New(getBlockchainConfig(cmd))
			if err != nil {
				return err
			}

			if err := n.Init(); err != nil {
				return err
			}

			w, err := n.GetNodeAccount()
			if err != nil {
				return err
			}

			filePath, _ := cmd.Flags().GetString("file")
			if path.Ext(filePath) != ".json" {
				return fmt.Errorf("file extension must be json")
			}

			if err := ioutil.WriteFile(filePath, w.KeyData, 0755); err != nil {
				return err
			}

			return writeOutput(cmd, map[string]interface{}{
				"address": w.Address(),
				"auth":    w.Auth,
			})
		},
		TraverseChildren: true,
	}

	addNodeIdFlag(nodeAccountDumpCmd)

	nodeAccountDumpCmd.Flags().StringP("file", "f", "", "Specify key file path to write")
	nodeAccountDumpCmd.MarkFlagRequired("file")

	return nodeAccountDumpCmd
}
