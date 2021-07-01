package node

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/rovergulf/rbn/wallets"
)

func (n *Node) SaveNodeAccount(w *wallets.Wallet) error {
	data, err := w.Serialize()
	if err != nil {
		return err
	}

	return n.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte("acc"), data)
	})
}

func (n *Node) GetNodeAccount() (*wallets.Wallet, error) {
	var w wallets.Wallet

	if err := n.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("acc"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return w.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &w, nil
}
