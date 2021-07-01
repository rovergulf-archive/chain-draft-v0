package wallets

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
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
	db     *badger.DB
	logger *zap.SugaredLogger
	quit   chan struct{}
}

// NewManager returns wallets Manager instance
func NewManager(opts params.Options) (*Manager, error) {
	badgerOpts := badger.DefaultOptions(opts.WalletsFilePath)
	db, err := badgerdb.OpenDB(opts.WalletsFilePath, badgerOpts)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	return &Manager{
		db:     db,
		logger: opts.Logger,
	}, err
}

func (m *Manager) DbSize() (int64, int64) {
	return m.db.Size()
}

func (m *Manager) Shutdown() {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			m.logger.Errorf("Unable to close wallets db: %s", err)
		}
	}
}

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
		Auth: auth,
		key:  key,
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

func (m *Manager) FindAccountKey(address common.Address) ([]byte, error) {
	var privateKey []byte
	if err := m.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(address.Bytes())
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			privateKey = val
			return nil
		})
	}); err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (m *Manager) GetWallet(address common.Address, auth string) (*Wallet, error) {
	encryptedKey, err := m.FindAccountKey(address)
	if err != nil {
		return nil, err
	}

	key, err := keystore.DecryptKey(encryptedKey, auth)
	if err != nil {
		return nil, err
	}

	return &Wallet{
		Auth: auth,
		key:  key,
	}, nil
}
