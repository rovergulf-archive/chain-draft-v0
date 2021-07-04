package core

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"time"
)

const (
	TxFee   = uint64(50)
	TxLimit = uint64(1 << 10)
)

var (
	txPrefix       = []byte("tx/")
	txPrefixLength = len(txPrefix)
)

// NewTransaction creates a new transaction
func NewTransaction(from, to common.Address, amount uint64, nonce uint64, nether, netherPrice uint64, data []byte) (types.Transaction, error) {
	if from == to {
		return types.Transaction{}, fmt.Errorf("transaction cannot be sent to yourself")
	}

	return types.Transaction{
		From:        from,
		To:          to,
		Value:       amount,
		Nonce:       nonce,
		Nether:      nether,
		NetherPrice: netherPrice,
		Data:        data,
		Time:        time.Now().Unix(),
	}, nil
}

func (bc *Blockchain) ListTransactions() ([]types.SignedTx, error) {
	var txs []types.SignedTx

	if err := bc.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.Prefix = txPrefix
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var tx types.SignedTx

			if err := item.Value(func(val []byte) error {
				return tx.Deserialize(val)
			}); err != nil {
				return err
			}

			txs = append(txs, tx)
		}
		return nil
	}); err != nil {
		bc.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return txs, nil
}

func (bc *Blockchain) FindTransaction(txId []byte) (*types.SignedTx, error) {
	var tx types.SignedTx

	if err := bc.db.View(func(txn *badger.Txn) error {
		key := append(txPrefix, txId...)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return tx.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &tx, nil
}

func (bc *Blockchain) SaveTx(txHash common.Hash, tx types.SignedTx) error {
	encodedTx, err := tx.Serialize()
	if err != nil {
		return err
	}

	key := append(txPrefix, txHash.Bytes()...)
	return bc.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, encodedTx)
	})
}

func (bc *Blockchain) ApplyTx(tx types.SignedTx) error {
	fromAddr, err := bc.GetBalance(tx.From)
	if err != nil {
		bc.logger.Errorf("Unable to get sender balance: %s", err)
		return err
	}

	toAddr, err := bc.GetBalance(tx.To)
	if err != nil {
		bc.logger.Errorf("Unable to get recipient balance: %s", err)
		return err
	}

	if tx.Cost() > fromAddr.Balance {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB",
			tx.From.String(), fromAddr.Balance, tx.Cost())
	}

	fromAddr.Balance -= tx.Cost()
	toAddr.Balance += tx.Value

	fromAddr.Nonce = tx.Nonce

	from, err := fromAddr.Serialize()
	if err != nil {
		return err
	}

	to, err := toAddr.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		senderKey := append(balancesPrefix, fromAddr.Address.Bytes()...)
		if err := txn.Set(senderKey, from); err != nil {
			return err
		}

		recipientKey := append(balancesPrefix, toAddr.Address.Bytes()...)
		return txn.Set(recipientKey, to)
	})
}
