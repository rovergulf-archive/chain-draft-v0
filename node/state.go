package node

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"sync"
)

// is a temporary state required for tx validations before block is mined
type pendingState struct {
	lock             *sync.RWMutex
	currentBlockHash common.Hash
	transactions     map[common.Hash]*types.SignedTx
	rejectedTxs      []*types.SignedTx
	balances         map[common.Address]*types.Balance
}

func newPendingState() *pendingState {
	return &pendingState{
		lock:         new(sync.RWMutex),
		transactions: make(map[common.Hash]*types.SignedTx),
		balances:     make(map[common.Address]*types.Balance),
	}
}

func (s *pendingState) reset() {
	s.transactions = make(map[common.Hash]*types.SignedTx)
	s.balances = make(map[common.Address]*types.Balance)
}

func (s *pendingState) pendingTxLen() int {
	var txLen int
	s.lock.RLock()
	txLen = len(s.transactions)
	s.lock.RUnlock()
	return txLen
}

func (s *pendingState) getTxsAsArray(limit int) []*types.SignedTx {
	var results []*types.SignedTx
	var txs map[common.Hash]*types.SignedTx
	s.lock.RLock()
	txs = s.transactions
	s.lock.RUnlock()

	for txHash := range txs {
		tx := txs[txHash]

		results = append(results, tx)

		if len(results) >= limit {
			break
		}
	}

	return results
}

func (s *pendingState) getTx(txHash common.Hash) (*types.SignedTx, bool) {
	var tx *types.SignedTx
	var ok bool
	s.lock.RLock()
	tx, ok = s.transactions[txHash]
	s.lock.RUnlock()
	return tx, ok
}

func (s *pendingState) addTx(txHash common.Hash, tx *types.SignedTx) error {
	s.lock.Lock()
	s.transactions[txHash] = tx
	s.lock.Unlock()
	return nil
}

func (s *pendingState) removeTx(txHash common.Hash) {
	s.lock.Lock()
	delete(s.transactions, txHash)
	s.lock.Unlock()
}

func (s *pendingState) getBalance(address common.Address) (*types.Balance, bool) {
	var b *types.Balance
	var ok bool
	s.lock.RLock()
	b, ok = s.balances[address]
	s.lock.RUnlock()
	return b, ok
}

func (s *pendingState) addBalance(address common.Address, balance *types.Balance) error {
	s.lock.Lock()
	s.balances[address] = balance
	s.lock.Unlock()
	return nil
}

func (s *pendingState) removeBalance(address common.Address) {
	s.lock.Lock()
	delete(s.balances, address)
	s.lock.Unlock()
}
