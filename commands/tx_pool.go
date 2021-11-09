package commands

import "github.com/rovergulf/chain/core/types"

// TBD or moved
type TxPool struct {
	pending types.Transactions
	queued  types.Transactions
}
