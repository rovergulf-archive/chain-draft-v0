package database

import "github.com/rovergulf/rbn/core"

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
	GetGenesis() (*core.Genesis, error)
	LatestBlock() (*core.Block, error)
	AddBlock(key string, block core.Block) error
	GetBlock(key string) (*core.Block, error)
	GetBlocks() ([]*core.Block, error)
	SaveTransaction(txHash []byte, data []byte) error
	FindTransaction(txHash []byte) (*core.SignedTx, error)
	ListTransactions() ([]*core.SignedTx, error)

	Shutdown()
}

// NodeBackend represents RBN node.Node backend
type NodeBackend interface {
	Shutdown()
}
