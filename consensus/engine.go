package consensus

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
)

// Engine is the consensus Node interface
type Engine interface {
	// Author returns a consensus game winner who would mine block and take the most of nether pool award
	Author(ctx context.Context) (common.Address, error)

	// VerifyHeader waits for required amount of network verifications
	VerifyHeader(header types.BlockHeader, sig []byte) error

	Prepare(ctx context.Context, block *types.Block) error

	Apply(ctx context.Context) error

	// Finalize applies post-transactions
	Finalize(ctx context.Context, transactions []*types.SignedTx, receipts []*types.Receipt) (*types.Block, error)
}
