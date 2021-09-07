package commands

import (
	"crypto/tls"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/params"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

var (
	cfgFile, dataDir string
	logger           *zap.SugaredLogger
	blockChain       *core.BlockChain
	accountManager   *wallets.Manager
	localNode        *node.Node
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "rbn",
	Short:   "Rovergulf BlockChain CLI",
	Long:    `Rovergulf BlockChain Network SDK`,
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// config
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rbn/config.yaml)")

	// logger & debug opts
	rootCmd.PersistentFlags().Bool("log_json", false, "Enable JSON formatted logs output")
	rootCmd.PersistentFlags().Int("log_level", int(zapcore.DebugLevel), "Log level")
	rootCmd.PersistentFlags().Bool("log_stacktrace", false, "Log stacktrace verbose")
	rootCmd.PersistentFlags().Bool("dev", false, "Enable development/testing environment")

	// main flags
	rootCmd.PersistentFlags().StringVar(&dataDir, "data_dir", os.Getenv("DATA_DIR"), "BlockChain data directory")
	rootCmd.PersistentFlags().StringVar(&dataDir, "network_id", params.MainNetworkId, "Chain network id")

	// bind viper values
	bindViperPersistentFlag(rootCmd, "network.id", "network_id")
	bindViperPersistentFlag(rootCmd, "app.dev", "dev")
	bindViperPersistentFlag(rootCmd, "log_json", "log_json")
	bindViperPersistentFlag(rootCmd, "log_level", "log_level")
	bindViperPersistentFlag(rootCmd, "log_stacktrace", "log_stacktrace")
	bindViperPersistentFlag(rootCmd, "data_dir", "data_dir")

	// show version
	rootCmd.Flags().BoolP("version", "v", false, "Display version")

	// other commands

	// balances
	rootCmd.AddCommand(balancesCmd)
	balancesCmd.AddCommand(balancesListCmd())
	balancesCmd.AddCommand(balancesGetCmd())

	// chain
	rootCmd.AddCommand(blockchainCmd())

	// node
	rootCmd.AddCommand(nodeCmd())

	// tx
	rootCmd.AddCommand(txCmd())

	// wallets
	rootCmd.AddCommand(walletsCmd())

	initZapLogger()
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in $HOME and /opt/rbn directory with name "config.yaml".
		viper.AddConfigPath(home)
		viper.AddConfigPath("/opt/rbn")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config.yaml")
	}

	setConfigDefaults()

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func setConfigDefaults() {
	viper.SetDefault("metrics", true)
	viper.SetDefault(traceutil.JaegerTraceConfigKey, os.Getenv("JAEGER_TRACE"))

	viper.SetDefault("node_id", "")

	// storage
	viper.SetDefault("db", "")
	viper.SetDefault("data_dir", "tmp")
	viper.SetDefault("keystore", "")
	viper.SetDefault("pid_file", "/var/run/rbn/pidfile")

	// TBD dgraph connection settings
	// !!! Database interface needs to be implemented to use that
	//viper.SetDefault("dgraph.enabled", false)
	//viper.SetDefault("dgraph.host", "127.0.0.1")
	//viper.SetDefault("dgraph.port", "9080")
	//viper.SetDefault("dgraph.user", "")
	//viper.SetDefault("dgraph.password", "")

	// ssl configuration
	viper.SetDefault("ssl.enabled", false)
	viper.SetDefault("ssl.email", "")
	viper.SetDefault("ssl.ca", "")
	viper.SetDefault("ssl.cert", "")
	viper.SetDefault("ssl.key", "")
	viper.SetDefault("ssl.verify", false)
	viper.SetDefault("ssl.mode", tls.NoClientCert)

	// chain network setup
	viper.SetDefault("network.id", params.MainNetworkId)
	viper.SetDefault("network.addr", "127.0.0.1:9420")
	viper.SetDefault("network.discovery", "swarm.rovergulf.net:443")

	// p2p settings
	viper.SetDefault("node.max_peers", 256)
	viper.SetDefault("node.addr", "127.0.0.1")
	viper.SetDefault("node.port", 9420)
	viper.SetDefault("node.sync_mode", node.SyncModeDefault)
	viper.SetDefault("node.sync_interval", 5)
	viper.SetDefault("node.cache_dir", "")
	viper.SetDefault("node.no_discovery", false)

	// http server
	viper.SetDefault("http.disabled", false)
	viper.SetDefault("http.addr", "127.0.0.1")
	viper.SetDefault("http.port", 9469)

	// TBD
	// Cache
	//viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.size", 256<<20) // 256mb

	// Runtime configuration
	//viper.SetDefault("runtime.max_cpu", runtime.NumCPU())
	//viper.SetDefault("runtime.max_mem", getAvailableOSMemory())

}

// initializes zap.SugaredLogger instance for logger
func initZapLogger() {
	cfg := zap.NewDevelopmentConfig()
	cfg.Development = viper.GetBool("dev")
	cfg.DisableStacktrace = !viper.GetBool("log_stacktrace")

	if logJson := viper.GetBool("log_json"); logJson {
		cfg.Encoding = "json"
	} else {
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logLevel, ok := viper.Get("log_level").(int)
	if !ok {
		logLevel = int(zapcore.DebugLevel)
	}

	cfg.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel))
	l, err := cfg.Build()
	if err != nil {
		log.Fatalf("Failed to run zap logger: %s", err)
	}

	logger = l.Sugar()
}
