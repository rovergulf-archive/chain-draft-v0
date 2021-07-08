package core

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/rovergulf/rbn/core/types"
	"github.com/spf13/viper"
)

func (bc *BlockChain) NewGenesisBlockWithRewrite(ctx context.Context) error {
	gen := genesisByNetworkId(viper.GetString("network.id"))

	genesisBlock, err := gen.ToBlock()
	if err != nil {
		bc.logger.Errorf("Unable to prepare genesis block")
		return err
	}

	genSerialized, err := gen.Serialize()
	if err != nil {
		bc.logger.Errorf("Unable to marshal genesis: %s", err)
		return err
	}

	serializedBLock, err := genesisBlock.Serialize()
	if err != nil {
		bc.logger.Errorf("Unable to serialize genesis block: %s", err)
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte("g"), genSerialized); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}
		if err := txn.Set([]byte("gen_block"), serializedBLock); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		if err := txn.Set(genesisBlock.BlockHeader.Hash.Bytes(), serializedBLock); err != nil {
			bc.logger.Errorf("Unable to put genesis block: %s", err)
			return err
		}

		if err := txn.Set([]byte("lh"), genesisBlock.BlockHeader.Hash.Bytes()); err != nil {
			bc.logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		for i := range genesisBlock.Transactions {
			tx := genesisBlock.Transactions[i]

			bal := types.Balance{
				Address: tx.To,
				Balance: tx.Value,
				Nonce:   0,
			}

			balanceEncoded, err := bal.Serialize()
			if err != nil {
				return err
			}

			balanceKey := append(balancesPrefix, tx.To.Bytes()...)
			if err := txn.Set(balanceKey, balanceEncoded); err != nil {
				bc.logger.Errorf("Unable to save balance: %s", err)
				return err
			}
		}

		bc.genesis = gen
		bc.LastHash = genesisBlock.BlockHeader.Hash
		return nil
	})
}

func (bc *BlockChain) loadGenesis(ctx context.Context) error {
	return bc.db.View(func(txn *badger.Txn) error {
		gen, err := txn.Get([]byte("g"))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrGenesisNotExists
			}
			bc.logger.Errorf("Unable to load genesis from storage: %s", err)
			return err
		}

		return gen.Value(func(val []byte) error {
			return bc.genesis.Deserialize(val)
		})
	})
}

func (bc *BlockChain) GetGenesis(ctx context.Context) (*Genesis, error) {
	if bc.genesis == nil {
		bc.genesis = new(Genesis)
		if err := bc.loadGenesis(ctx); err != nil {
			return nil, err
		}
	}
	return bc.genesis, nil
}
