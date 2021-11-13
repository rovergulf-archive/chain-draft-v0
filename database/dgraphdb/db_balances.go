package dgraphdb

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/chain/core/types"
)

func (bc *chainDb) NewBalance(ctx context.Context, b *types.Balance) error {
	return nil
}

func (bc *chainDb) UpdateBalances(ctx context.Context, b []*types.Balance) error {
	return nil
}

func (bc *chainDb) GetBalance(ctx context.Context, addr common.Address) (*types.Balance, error) {
	var b types.Balance
	return &b, nil
}

func (bc *chainDb) SearchBalances(ctx context.Context) ([]*types.Balance, error) {
	var results []*types.Balance
	return results, nil
}
