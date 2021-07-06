package database

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/wallets"
)

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

// ChainBackend represents multiple drivers storage interface for core.BlockChain
type ChainBackend interface {
	SaveGenesis(ctx context.Context, genesis *core.Genesis) error
	GetGenesis(ctx context.Context) (*core.Genesis, error)

	NewBalance(ctx context.Context, balance *types.Balance) error
	UpdateBalance(ctx context.Context, balance *types.Balance) error
	GetBalance(ctx context.Context, hash common.Address) (*types.Balance, error)
	SearchBalances(ctx context.Context) ([]*types.Balance, error)

	NewBlockHeader(ctx context.Context, block *types.BlockHeader) error
	GetBlockHeader(ctx context.Context, hash common.Hash) (*types.BlockHeader, error)
	SearchBlockHeaders(ctx context.Context) ([]*types.BlockHeader, error)

	NewBlock(ctx context.Context, block *types.Block) error
	GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error)
	SearchBlocks(ctx context.Context) ([]*types.Block, error)

	NewTransaction(ctx context.Context, tx *types.Transaction) error
	RemoveTransaction(ctx context.Context, txHash common.Hash) error
	GetTransaction(ctx context.Context, txHash common.Hash) (*types.Transaction, error)
	SearchTransactions(ctx context.Context) ([]*types.Transaction, error)
}

// KeystoreBackend represents account's private key storage
type KeystoreBackend interface {
	NewAccountKey(ctx context.Context, address common.Address, encryptedKey []byte) error
	FindAccountKey(ctx context.Context, address common.Address) ([]byte, error)
}

// NodeBackend represents RBN node.Node backend
type NodeBackend interface {
	NewNodeAccount(ctx context.Context, wallet *wallets.Wallet) error
	GetNodeAccount(ctx context.Context) (*wallets.Wallet, error)

	NewPeerNode(ctx context.Context) error
	GetPeerNode(ctx context.Context) error
	SearchPeerNodes(ctx context.Context) error

	// TODO: network and node stats
}
