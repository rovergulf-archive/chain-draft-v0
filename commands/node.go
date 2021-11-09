package commands

import (
	"context"
	"fmt"
	"github.com/rovergulf/chain/node"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

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
	//nodeCmd.AddCommand(nodeAccountImportCmd())

	return nodeCmd
}

func nodeRunCmd() *cobra.Command {
	var nodeRunCmd = &cobra.Command{
		Use:   "run",
		Short: "Run Rovergulf BlockChain Network peer node",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			n, err := node.New(getBlockChainConfig(cmd))
			if err != nil {
				return err
			}

			if err := n.Init(ctx); err != nil {
				return err
			}

			return n.Run(ctx)
		},
		TraverseChildren: true,
	}

	// gRPC
	nodeRunCmd.Flags().String("node-addr", "127.0.0.1", "Node address would listen to")
	bindViperFlag(nodeRunCmd, "node.addr", "node-addr")
	nodeRunCmd.Flags().Int("node-port", 9420, "Node port would listen to accept gRPC connections")
	bindViperFlag(nodeRunCmd, "node.port", "node-port")
	// HTTP REST
	nodeRunCmd.Flags().String("http-addr", "127.0.0.1", "Node address would listen to")
	bindViperFlag(nodeRunCmd, "http.addr", "http-addr")
	nodeRunCmd.Flags().Int("http-port", 9469, "Node port would listen to accept Web API Requests")
	bindViperFlag(nodeRunCmd, "http.port", "http-port")
	// JSONRpc 2.0 – TBD (??)
	//nodeRunCmd.Flags().String("jrpc-addr", "127.0.0.1", "Node address would listen to")
	//bindViperFlag(nodeRunCmd, "jrpc.addr", "jrpc-addr")
	//nodeRunCmd.Flags().Int("jrpc-port", 9300, "Node port for JSON Rpc 2.0 Interface")
	//bindViperFlag(nodeRunCmd, "jrpc.port", "jrpc-port")

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

	return nodeStopCmd
}

func nodeAccountDumpCmd() *cobra.Command {
	nodeAccountDumpCmd := &cobra.Command{
		Use:   "account-dump",
		Short: "Export node account key",
		Long: `Exports node default account CryptJSON to specified file 
and prints out mnemonic passphrase to unlock it`,
		PreRunE: prepareNode,
		RunE: func(cmd *cobra.Command, args []string) error {
			//ctx := context.Background()
			//ctx, cancel := context.WithCancel(ctx)
			//defer cancel()
			defer localNode.Shutdown()

			w, err := localNode.GetNodeAccount()
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

	nodeAccountDumpCmd.Flags().StringP("file", "f", "", "Specify key file path to write")
	nodeAccountDumpCmd.MarkFlagRequired("file")

	return nodeAccountDumpCmd
}

func nodeAccountImportCmd() *cobra.Command {
	nodeAccountImportCmd := &cobra.Command{
		Use:   "account-import",
		Short: "Import node account key",
		Long: `Imports node account key from specified CryptoJSON file
and sets it as node default`,
		PreRunE: prepareNode,
		RunE: func(cmd *cobra.Command, args []string) error {
			//ctx, cancel := context.WithCancel(context.Background())
			//defer cancel()
			defer localNode.Shutdown()

			// TODO

			return writeOutput(cmd, map[string]interface{}{
				"address": "w.Address()",
				"auth":    "w.Auth",
			})
		},
		TraverseChildren: true,
	}

	nodeAccountImportCmd.Flags().StringP("file", "f", "", "Specify key file path to read")
	nodeAccountImportCmd.MarkFlagRequired("file")

	return nodeAccountImportCmd
}
