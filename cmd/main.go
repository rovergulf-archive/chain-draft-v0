// Copyright 2021 The rbn Authors
// This file is part of the rbn library.
//
// The rbn library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The rbn library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the rbn library. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/rovergulf/chain/core"
	"github.com/rovergulf/chain/node"
	"github.com/rovergulf/chain/params"
	"github.com/rovergulf/chain/pkg/configutil"
	"github.com/rovergulf/chain/wallets"
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

func main() {
	Execute()
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
	rootCmd.PersistentFlags().Int64("network_id", int64(params.MainNetworkId), "Chain network id")

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

	configutil.SetDefaultConfigValues()

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
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
