package commands

import (
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

var (
	cfgFile string
	logger  *zap.SugaredLogger
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.rnt.yaml)")
	rootCmd.PersistentFlags().Bool("log_json", false, "Enable JSON formatted logs output")
	rootCmd.PersistentFlags().Int("log_level", int(zapcore.DebugLevel), "Log level")
	rootCmd.PersistentFlags().String("data_dir", "tmp", "Blockchain data directory")

	bindViperPersistentFlag(rootCmd, "jaeger_trace_url", "jaeger_trace")
	bindViperPersistentFlag(rootCmd, "log_json", "log_json")
	bindViperPersistentFlag(rootCmd, "log_level", "log_level")
	bindViperPersistentFlag(rootCmd, "data_dir", "data_dir")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
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

		// Search config in home directory with name "config.yaml".
		viper.AddConfigPath(home)
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
	viper.SetDefault("metrics_port", 8080)
	viper.SetDefault("jaeger_trace", os.Getenv("JAEGER_TRACE"))

	viper.SetDefault("data_dir", "tmp")
	viper.SetDefault("root", "0x09ee50f2f37fcba1845de6fe5c762e83e65e755c")
	viper.SetDefault("miner", "0x0000000000000000000000000000000000000000")

	// ssl configuration
	viper.SetDefault("ssl.enabled", false)
	viper.SetDefault("ssl.email", "")
	viper.SetDefault("ssl.ca", "")
	viper.SetDefault("ssl.cert", "")
	viper.SetDefault("ssl.key", "")
	viper.SetDefault("ssl.verify", false)

	// bootstrap server
	viper.SetDefault("bootstrap.addr", "0.0.0.0") // chain.rovergulf.net
	viper.SetDefault("bootstrap.port", 9420)

	// http server
	viper.SetDefault("http.addr", "0.0.0.0")
	viper.SetDefault("http.port", 9069)

	// TBD
	// Runtime configuration
	//viper.SetDefault("runtime.max_cpu", runtime.NumCPU())
	//viper.SetDefault("runtime.max_mem", getAvailableOSMemory())

}

// initializes zap.SugaredLogger instance for logger
func initZapLogger() {
	config := zap.NewDevelopmentConfig()
	config.Development = viper.GetBool("dev")
	config.DisableStacktrace = viper.GetBool("log_stacktrace")

	if logJson := viper.GetBool("log_json"); logJson {
		config.Encoding = "json"
	} else {
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logLevel, ok := viper.Get("log_level").(int)
	if !ok {
		logLevel = int(zapcore.DebugLevel)
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.Level(logLevel))
	l, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to run zap logger: %s", err)
	}

	logger = l.Sugar()
	viper.Set("logger", logger)
}
