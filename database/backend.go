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

// ChainBackend represents multiple drivers storage interface for core.Blockchain
type ChainBackend interface {
}

// KeystoreBackend represents account's private key storage
type KeystoreBackend interface {
}

// NodeBackend represents RBN node.Node backend
type NodeBackend interface {
}
