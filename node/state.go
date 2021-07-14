package node

import (
	"errors"
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
	receipts         map[common.Hash]*types.Receipt
}

func newPendingState() *pendingState {
	return &pendingState{
		lock:         new(sync.RWMutex),
		transactions: make(map[common.Hash]*types.SignedTx),
		balances:     make(map[common.Address]*types.Balance),
		receipts:     make(map[common.Hash]*types.Receipt),
	}
}

func (s *pendingState) reset() {
	s.transactions = make(map[common.Hash]*types.SignedTx)
	s.balances = make(map[common.Address]*types.Balance)
	s.receipts = make(map[common.Hash]*types.Receipt)
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

func (s *pendingState) getReceipt(hash common.Hash) (*types.Receipt, bool) {
	var r *types.Receipt
	var ok bool
	s.lock.RLock()
	r, ok = s.receipts[hash]
	s.lock.RUnlock()
	return r, ok
}

func (s *pendingState) addReceipt(hash common.Hash, balance *types.Receipt) error {
	s.lock.Lock()
	s.receipts[hash] = balance
	s.lock.Unlock()
	return nil
}

func (s *pendingState) removeReceipt(hash common.Hash) {
	s.lock.Lock()
	delete(s.receipts, hash)
	s.lock.Unlock()
}

func (s *pendingState) applyPendingTx(txHash common.Hash) (*types.Receipt, error) {
	tx, ok := s.getTx(txHash)
	if !ok {
		return nil, errors.New("pending tx not found")
	}

	sender, ok := s.getBalance(tx.From)
	if !ok {
		return nil, errors.New("pending sender balance not found")
	}

	recipient, ok := s.getBalance(tx.From)
	if !ok {
		return nil, errors.New("pending recipient balance not found")
	}

	sender.Balance -= tx.Cost()
	recipient.Balance += tx.Value

	sender.Nonce = tx.Nonce
	recipient.Nonce++

	receipt := &types.Receipt{
		Addr:        sender.Address,
		Balance:     sender.Balance,
		NetherUsed:  tx.Nether,
		NetherPrice: tx.NetherPrice,
		TxHash:      txHash,
	}

	if err := s.addBalance(sender.Address, sender); err != nil {
		return nil, err
	}

	if err := s.addBalance(recipient.Address, recipient); err != nil {
		return nil, err
	}

	if err := s.addReceipt(txHash, receipt); err != nil {
		return nil, err
	}

	return receipt, nil
}
