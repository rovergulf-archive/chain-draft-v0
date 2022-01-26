package dgraphdb

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/chain/core/types"
)

func (bc *chainDb) NewBlockHeader(ctx context.Context, h *types.BlockHeader) error {
	return nil
}

func (bc *chainDb) GetBlockHeader(ctx context.Context, hash common.Hash) (*types.BlockHeader, error) {
	var h types.BlockHeader
	return &h, nil
}

func (bc *chainDb) SearchBlockHeaders(ctx context.Context) ([]*types.BlockHeader, error) {
	var results []*types.BlockHeader
	return results, nil
}

func (bc *chainDb) NewBlock(ctx context.Context, b *types.Block) error {
	return nil
}

func (bc *chainDb) GetBlock(ctx context.Context, hash common.Hash) (*types.Block, error) {
	var b types.Block
	return &b, nil
}

func (bc *chainDb) SearchBlocks(ctx context.Context) ([]*types.Block, error) {
	var results []*types.Block
	return results, nil
}
