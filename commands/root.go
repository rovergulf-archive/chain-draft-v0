package commands

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"os"
)

var (
	cfgFile, dataDir string
	logger           *zap.SugaredLogger
	//bc *core.Blockchain
	//node *node.Node
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "rbn",
	Short:   "Rovergulf Blockchain CLI",
	Long:    `Rovergulf Blockchain Network SDK`,
	Version: "0.0.1-dev",
	RunE: func(cmd *cobra.Command, args []string) error {
		ver, _ := cmd.Flags().GetBool("version")
		if ver {
			return writeOutput(cmd, cmd.Version)
		} else {
			return cmd.Usage()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
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

	// main flags
	rootCmd.PersistentFlags().StringVar(&dataDir, "data_dir", os.Getenv("DATA_DIR"), "Blockchain data directory")

	// bind viper values
	bindViperPersistentFlag(rootCmd, "log_json", "log_json")
	bindViperPersistentFlag(rootCmd, "log_level", "log_level")
	bindViperPersistentFlag(rootCmd, "log_stacktrace", "log_stacktrace")
	bindViperPersistentFlag(rootCmd, "data_dir", "data_dir")

	// show version
	rootCmd.Flags().BoolP("version", "v", false, "Display version")

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
	viper.SetDefault("jaeger_trace", os.Getenv("JAEGER_TRACE"))

	viper.SetDefault("node_id", "")

	// node
	viper.SetDefault("root", "0x59fc6df01d2e84657faba24dc96e14871192bda4")
	viper.SetDefault("miner", "0x0000000000000000000000000000000000000000")

	// storage
	viper.SetDefault("db", "")
	viper.SetDefault("data_dir", "tmp")
	viper.SetDefault("backup_dir", backupsDir)

	viper.SetDefault("dgraph.enabled", false)
	viper.SetDefault("dgraph.host", "127.0.0.1")
	viper.SetDefault("dgraph.port", "9080")
	viper.SetDefault("dgraph.user", "")
	viper.SetDefault("dgraph.password", "")

	// ssl configuration
	viper.SetDefault("ssl.enabled", false)
	viper.SetDefault("ssl.email", "")
	viper.SetDefault("ssl.ca", "")
	viper.SetDefault("ssl.cert", "")
	viper.SetDefault("ssl.key", "")
	viper.SetDefault("ssl.verify", false)

	// bootstrap server
	viper.SetDefault("network.id", 1)                  //
	viper.SetDefault("network.host", "127.0.0.1:9420") // swarm.rovergulf.net:443

	// http server
	viper.SetDefault("node.addr", "127.0.0.1")
	viper.SetDefault("node.port", 9420)
	viper.SetDefault("node.sync_interval", 5)
	viper.SetDefault("node.cache_dir", "")

	viper.SetDefault("http.addr", "127.0.0.1")
	viper.SetDefault("http.port", 9069)

	// TBD
	// Runtime configuration
	//viper.SetDefault("runtime.max_cpu", runtime.NumCPU()) // take 2/3 available by default
	//viper.SetDefault("runtime.max_mem", getAvailableOSMemory()) // same as above

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
	viper.Set("logger", logger)
}

func initOpentracing(address string) (opentracing.Tracer, io.Closer, error) {
	metrics := prometheus.New()

	traceTransport, err := jaeger.NewUDPTransport(address, 0)
	if err != nil {
		logger.Errorf("Unable to setup tracing agent connection: %s", err)
		return nil, nil, err
	}

	tracer, closer, err := config.Configuration{
		ServiceName: "rbn",
	}.NewTracer(
		config.Sampler(jaeger.NewConstSampler(true)),
		config.Reporter(jaeger.NewRemoteReporter(
			traceTransport,
			jaeger.ReporterOptions.Logger(jaeger.StdLogger)),
		),
		config.Metrics(metrics),
	)
	if err != nil {
		logger.Errorf("Unable to start tracer: %s", err)
		return nil, nil, err
	}

	logger.Debugw("Jaeger tracing client initialized", "collector_url", address)
	return tracer, closer, nil
}
