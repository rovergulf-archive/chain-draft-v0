package database

import "context"

const (
	Badger = "badger" // idk, pretty sure etcd, bbolt or consul would fit as well
	Dgraph = "dgraph" // would research and implement: https://github.com/rovergulf/rbn/issues/28
)

// Config is used to configure and run RBN database backend
type Config struct {
	// Driver represents database interface, by default it is BadgerDB
	// value 'dgraph' may be specified to use Dgraph backends
	Driver string `json:"driver" yaml:"driver"`

	Dir  string `json:"dir" yaml:"dir"`
	Addr string `json:"addr" yaml:"addr"`
}

// lifecycle handles start/stop signals for db connection/file read
type lifecycle interface {
	//Run() error
	Shutdown()
}

type KvStorage interface {
	Get(ctx context.Context, key []byte) ([]byte, error)
	Put(ctx context.Context, key []byte, data []byte) error
	Delete(ctx context.Context, key []byte) error
	List(ctx context.Context, prefix []byte) ([][]byte, error)
}

type txFunc func(txn KvStorage) error

type kvTx interface {
	View(ctx context.Context, txFunc txFunc) error
	Update(ctx context.Context, txFunc txFunc) error
}

type Backend interface {
	lifecycle
	KvStorage
}
