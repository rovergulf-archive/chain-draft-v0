package dgraphdb

import (
	"context"
	"github.com/rovergulf/chain/core"
	"github.com/rovergulf/chain/core/types"
)

func (bc *chainDb) SaveGenesis(ctx context.Context, tx *core.Genesis) error {
	return nil
}

func (bc *chainDb) GetGenesis(ctx context.Context) (*core.Genesis, error) {
	var gen core.Genesis
	return &gen, nil
}

func (bc *chainDb) SaveGenesisBlock(ctx context.Context, tx *types.Block) error {
	return nil
}

func (bc *chainDb) GetGenesisBlock(ctx context.Context) (*types.Block, error) {
	var block types.Block
	return &block, nil
}
