package core

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"strconv"
)

const (
	DbFileName = "chain.db"
)

var (
	emptyHash = common.HexToHash("")
)

var (
	ErrGenesisNotExists     = errors.New("genesis does not exists")
	ErrBalanceNotExists     = errors.New("balance does not exists")
	ErrBalanceAlreadyExists = errors.New("balance already exists")
	ErrBlockNotExists       = errors.New("block does not exists")
	ErrBlockAlreadyExists   = errors.New("block already exists")
	ErrTxNotExists          = errors.New("transaction does not exists")
	ErrTxAlreadyExists      = errors.New("transaction already exists")
	ErrInvalidRewardData    = errors.New("invalid reward tx data")
	ErrReceiptNotExists     = errors.New("receipt does not exists")
	ErrReceiptAlreadyExists = errors.New("receipt already exists")
)

var (
	lastHashKey        = []byte("lh")
	genesisKey         = []byte("gen")
	genesisBlockKey    = []byte("root")
	blocksPrefix       = []byte("blocks/")
	blockNumbersPrefix = []byte("blockNums/")
	blockHeadersPrefix = []byte("headers/")
	balancesPrefix     = []byte("balances/")
	txsPrefix          = []byte("txs/")
	receiptsPrefix     = []byte("receipts/")
)

func blockDbPrefix(hash common.Hash) []byte {
	return append(blocksPrefix, hash.Bytes()...)
}

func blockNumDbPrefix(number uint64) []byte {
	numStr := strconv.FormatUint(number, 10)
	prefix := []byte(numStr)
	return append(blockNumbersPrefix, prefix...)
}

func blockHeaderDbPrefix(hash common.Hash) []byte {
	return append(blockHeadersPrefix, hash.Bytes()...)
}

func balanceDbPrefix(addr common.Address) []byte {
	return append(balancesPrefix, addr.Bytes()...)
}

func txDbPrefix(hash common.Hash) []byte {
	return append(txsPrefix, hash.Bytes()...)
}

func receiptDbPrefix(hash common.Hash) []byte {
	return append(receiptsPrefix, hash.Bytes()...)
}

func IsHashEmpty(hash common.Hash) bool {
	return bytes.Compare(hash.Bytes(), emptyHash.Bytes()) == 0
}
