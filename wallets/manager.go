package wallets

import (
	"bytes"
	"context"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/pkg/config"
	"go.uber.org/zap"
)

const DbWalletFile = "wallets.db"

type Backend interface {
	Put(ctx context.Context, key []byte, data []byte)
	Get(ctx context.Context, key []byte)
	List(ctx context.Context, prefix []byte)
	Delete(ctx context.Context, key []byte)
}

type Manager struct {
	Db *badger.DB `json:"-" yaml:"-"`

	backend Backend

	logger *zap.SugaredLogger
	quit   chan struct{}
}

// NewManager returns wallets Manager instance
func NewManager(opts config.Options) (*Manager, error) {
	badgerOpts := badger.DefaultOptions(opts.WalletsFilePath)
	db, err := badgerdb.OpenDB(opts.WalletsFilePath, badgerOpts)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	return &Manager{
		Db:     db,
		logger: opts.Logger,
	}, err
}

func (m *Manager) Shutdown() {
	if m.Db != nil {
		if err := m.Db.Close(); err != nil {
			m.logger.Errorf("Unable to close wallets db: %s", err)
		}
	}
}

func (m *Manager) AddWallet(auth string) (*Wallet, error) {
	key, err := NewRandomKey()
	if err != nil {
		return nil, err
	}

	encryptedKey, err := keystore.EncryptKey(key, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	if err := m.Db.Update(func(txn *badger.Txn) error {
		return txn.Set(key.Address.Bytes(), encryptedKey)
	}); err != nil {
		return nil, err
	}

	wallet := &Wallet{
		Address: key.Address,
	}

	return wallet, nil
}

func (m *Manager) GetAllAddresses() ([]common.Address, error) {
	var addresses []common.Address

	if err := m.Db.View(func(txn *badger.Txn) error {
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

func (m Manager) GetWallet(address common.Address) (*Wallet, error) {
	var w *Wallet

	if err := m.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(address.Bytes())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			wallet, err := DeserializeWallet(val)
			if err != nil {
				return err
			}

			w = wallet
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return w, nil
}

func DeserializeManager(data []byte) (map[string]Wallet, error) {
	wallets := make(map[string]Wallet)

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&wallets); err != nil {
		return nil, err
	}

	return wallets, nil
}
