package node

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/rovergulf/rbn/wallets"
)

func (n *Node) saveNodeAccount(w *wallets.Wallet) error {
	if err := w.EncryptKey(); err != nil {
		return err
	}

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

	if err := w.Open(); err != nil {
		return nil, err
	}

	return &w, nil
}

func (n *Node) setupNodeAccount() error {
	w, err := n.GetNodeAccount()
	if err != nil {
		if err == badger.ErrKeyNotFound {
			passphrase, err := wallets.NewRandomMnemonic()
			if err != nil {
				return err
			}
			key, err := wallets.NewRandomKey()
			if err != nil {
				return err
			}
			newWallet, err := n.wm.AddWallet(key, passphrase)
			if err != nil {
				return err
			}

			if err := n.saveNodeAccount(newWallet); err != nil {
				return err
			}
			w = newWallet
		}
		return err
	}

	n.account = w
	return nil
}
