package configutil

import (
	"github.com/rovergulf/chain/node"
	"github.com/rovergulf/chain/params"
	"github.com/rovergulf/chain/pkg/traceutil"
	"github.com/spf13/viper"
	"os"
)

func SetDefaultConfigValues() {
	viper.SetDefault("metrics", true)
	viper.SetDefault(traceutil.JaegerTraceConfigKey, os.Getenv("JAEGER_TRACE"))

	// storage
	viper.SetDefault("db", "")
	viper.SetDefault("data_dir", "tmp")
	viper.SetDefault("keystore", "")

	// process id
	viper.SetDefault("pid_file", "/var/run/rbn/pidfile")

	// TBD dgraph connection settings
	// !!! Database interface needs to be implemented to use that
	viper.SetDefault("dgraph.enabled", false)
	viper.SetDefault("dgraph.host", "127.0.0.1")
	viper.SetDefault("dgraph.port", "9080")
	viper.SetDefault("dgraph.user", "")
	viper.SetDefault("dgraph.password", "")
	viper.SetDefault("dgraph.tls.enabled", false)
	viper.SetDefault("dgraph.tls.cert", "")
	viper.SetDefault("dgraph.tls.key", "")
	viper.SetDefault("dgraph.tls.verify", false)
	viper.SetDefault("dgraph.tls.auth", "")

	// chain network setup
	viper.SetDefault("network.id", params.MainNetworkId)

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
	viper.SetDefault("http.dial_timeout", 30)
	viper.SetDefault("http.read_timeout", 30)
	viper.SetDefault("http.write_timeout", 30)
	viper.SetDefault("http.ssl.enabled", false)
	viper.SetDefault("http.ssl.cert", "")
	viper.SetDefault("http.ssl.key", "")
	viper.SetDefault("http.ssl.verify", false)

	// TBD
	// Cache
	//viper.SetDefault("cache.enabled", false)
	viper.SetDefault("cache.size", 256<<20) // 256mb

	// Runtime configuration
	//viper.SetDefault("runtime.max_cpu", runtime.NumCPU())
	//viper.SetDefault("runtime.max_mem", getAvailableOSMemory())
}
