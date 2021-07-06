package badgerdb

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/database"
)

var (
	genesisDbKey      = []byte("genesis")
	genesisBlockDbKey = []byte("gen_block")
)

func (bc *chainDb) SaveGenesis(ctx context.Context, genesis *core.Genesis) error {
	genSerialized, err := genesis.Serialize()
	if err != nil {
		bc.logger.Errorf("Unable to serialize genesis: %s", err)
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(genesisDbKey, genSerialized); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		return nil
	})
}

func (bc *chainDb) GetGenesis(ctx context.Context) (*core.Genesis, error) {
	var gen core.Genesis
	if err := bc.db.View(func(txn *badger.Txn) error {
		lh, err := txn.Get(genesisDbKey)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return database.ErrGenesisNotExists
			}
			bc.logger.Errorf("Unable to load genesis from storage: %s", err)
			return err
		}

		return lh.Value(func(val []byte) error {
			return gen.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &gen, nil
}

func (bc *chainDb) SaveGenesisBlock(ctx context.Context, genesisBlock *types.Block) error {
	genSerialized, err := genesisBlock.Serialize()
	if err != nil {
		bc.logger.Errorf("Unable to serialize genesis: %s", err)
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(genesisBlockDbKey, genSerialized); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		return nil
	})
}

func (bc *chainDb) GetGenesisBlock(ctx context.Context) (*types.Block, error) {
	var b types.Block
	if err := bc.db.View(func(txn *badger.Txn) error {
		lh, err := txn.Get(genesisBlockDbKey)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return database.ErrBlockNotExists
			}
			bc.logger.Errorf("Unable to load genesis block from storage: %s", err)
			return err
		}

		return lh.Value(func(val []byte) error {
			return b.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &b, nil
}
