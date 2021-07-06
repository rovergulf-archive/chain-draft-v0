package badgerdb

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

var (
	txPrefix       = []byte("tx/")
	txPrefixLength = len(txPrefix)
)

func (bc *chainDb) NewTransaction(ctx context.Context, tx *types.Transaction) error {
	return nil
}

func (bc *chainDb) RemoveTransaction(ctx context.Context, hash common.Hash) error {
	return nil
}

func (bc *chainDb) GetTransaction(ctx context.Context, hash common.Hash) (*types.Transaction, error) {
	var tx types.Transaction
	return &tx, nil
}

func (bc *chainDb) SearchTransactions(ctx context.Context) ([]*types.Transaction, error) {
	var txs []*types.Transaction
	return txs, nil
}
