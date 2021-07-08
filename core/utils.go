package core

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
)

const (
	DbFileName = "chain.db"
)

var (
	emptyHash = common.HexToHash("")
)

var (
	ErrGenesisNotExists = errors.New("genesis does not exists")
	ErrBlockNotExists   = errors.New("block does not exists")
	ErrTxNotExists      = errors.New("transaction does not exists")
)

var (
	lastHashKey        = []byte("lh")
	genesisKey         = []byte("gen")
	genesisBlockKey    = []byte("root")
	blocksPrefix       = []byte("blocks/")
	blockHeadersPrefix = []byte("headers/")
	balancesPrefix     = []byte("balances/")
	txPrefix           = []byte("tx/")
)
