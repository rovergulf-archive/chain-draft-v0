package dgraphdb

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
)

func (bc *chainDb) GetLastHash(ctx context.Context) (common.Hash, error) {
	return common.Hash{}, nil
}
