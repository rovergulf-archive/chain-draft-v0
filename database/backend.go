package database

const (
	Badger = "badger"
	Dgraph = "dgraph"
)

// Config is used to configure and run RBN database backend
type Config struct {
	// Driver represents database interface, by default it is BadgerDB
	// value 'dgraph' may be specified to use Dgraph backends
	Driver string `json:"driver" yaml:"driver"`

	Dir  string `json:"dir" yaml:"dir"`
	Addr string `json:"addr" yaml:"addr"`
}

type DbStats struct {
	Size  int64                  `json:"size,omitempty" yaml:"size,omitempty"`
	Sizes map[string]interface{} `json:"sizes,omitempty" yaml:"sizes,omitempty"`
}

type Backend interface {
	lifecycle
}

type lifecycle interface {
	Shutdown()
}
