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
		n.logger.Errorf("Unable to serialize node account: %s", err)
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
			if err != badger.ErrKeyNotFound {
				n.logger.Errorf("Unable to get node account: %s", err)
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return w.Deserialize(val)
		})
	}); err != nil {
		if err != badger.ErrKeyNotFound {
			n.logger.Errorf("Unable to begin database read transaction: %s", err)
		}
		return nil, err
	}

	if err := w.Open(); err != nil {
		n.logger.Errorf("Unable to open node account key: %s", err)
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
				n.logger.Errorf("Unable to generate mnemonic pasphrase")
				return err
			}

			key, err := wallets.NewRandomKey()
			if err != nil {
				return err
			}

			newWallet, err := n.wm.AddWallet(key, passphrase)
			if err != nil {
				n.logger.Errorf("Unable to save node account to keystore: %s", err)
				return err
			}

			if _, err := n.bc.NewBalance(newWallet.Address(), 0); err != nil {
				n.logger.Errorf("Unable to create node balance")
				return err
			}

			if err := n.saveNodeAccount(newWallet); err != nil {
				n.logger.Errorf("Unable to save node account to node storage: %s", err)
				return err
			}

			w = newWallet
		} else {
			return err
		}
	}

	n.account = w
	return nil
}
