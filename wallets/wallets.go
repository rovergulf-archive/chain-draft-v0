package wallets

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
)

func (m *Manager) AddWallet(key *keystore.Key, auth string) (*Wallet, error) {
	encryptedKey, err := keystore.EncryptKey(key, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	if err := m.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key.Address.Bytes(), encryptedKey)
	}); err != nil {
		return nil, err
	}

	wallet := &Wallet{
		Auth:    auth,
		KeyData: encryptedKey,
		key:     key,
	}

	return wallet, nil
}

func (m *Manager) GetAllAddresses() ([]common.Address, error) {
	var addresses []common.Address

	if err := m.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			addresses = append(addresses, common.BytesToAddress(item.Key()))
		}
		return nil
	}); err != nil {
		m.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return addresses, nil
}

func (m *Manager) findAccountKey(address common.Address) ([]byte, error) {
	var privateKey []byte
	if err := m.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(address.Bytes())
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrAccountNotExists
			}
			return err
		}

		return item.Value(func(val []byte) error {
			privateKey = append(privateKey, val...)
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (m *Manager) GetWallet(address common.Address, auth string) (*Wallet, error) {
	encryptedKey, err := m.findAccountKey(address)
	if err != nil {
		return nil, err
	}

	key, err := keystore.DecryptKey(encryptedKey, auth)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Auth:    auth,
		KeyData: encryptedKey,
		key:     key,
	}, nil
}

func (m *Manager) Exists(ctx context.Context, address common.Address) error {
	return m.db.View(func(txn *badger.Txn) error {
		if _, err := txn.Get(address.Bytes()); err != nil {
			return err
		} else {
			return nil
		}
	})
}
