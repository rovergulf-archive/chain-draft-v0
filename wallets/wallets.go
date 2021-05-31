package wallets

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/rovergulf/rbn/pkg/repo"
	"go.uber.org/zap"
)

const DbWalletFile = "wallets.db"

type Wallets struct {
	Db     *badger.DB `json:"-" yaml:"-"`
	logger *zap.SugaredLogger
}

func InitWallets(opts config.Options) (*Wallets, error) {
	badgerOpts := badger.DefaultOptions(opts.WalletsFilePath)
	db, err := repo.OpenDB(opts.WalletsFilePath, badgerOpts)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	return &Wallets{
		Db:     db,
		logger: opts.Logger,
	}, err
}

func (ws *Wallets) Shutdown() {
	if ws.Db != nil {
		if err := ws.Db.Close(); err != nil {
			ws.logger.Errorf("Unable to close wallets db: %s", err)
		}
	}
}

func (ws *Wallets) AddWallet(auth string) (*Wallet, error) {
	key, err := NewRandomKey()
	if err != nil {
		return nil, err
	}

	encryptedKey, err := keystore.EncryptKey(key, auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}

	if err := ws.Db.Update(func(txn *badger.Txn) error {
		return txn.Set(key.Address.Bytes(), encryptedKey)
	}); err != nil {
		return nil, err
	}

	wallet := &Wallet{
		Address: key.Address,
	}

	return wallet, nil
}

func (ws *Wallets) GetAllAddresses() ([]common.Address, error) {
	var addresses []common.Address

	if err := ws.Db.View(func(txn *badger.Txn) error {
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
		ws.logger.Errorw("Unable to iterate db view", "err", err)
		return nil, err
	}

	return addresses, nil
}

func (ws Wallets) GetWallet(address common.Address) (*Wallet, error) {
	var w *Wallet

	if err := ws.Db.View(func(txn *badger.Txn) error {
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

func DeserializeWallets(data []byte) (map[string]Wallet, error) {
	wallets := make(map[string]Wallet)

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&wallets); err != nil {
		return nil, err
	}

	return wallets, nil
}
