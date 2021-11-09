package core

import (
	"context"
	"github.com/dgraph-io/badger/v3"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/chain/core/types"
	"github.com/rovergulf/chain/pkg/traceutil"
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

	blockKey := blockDbPrefix(genesisBlock.BlockHeader.BlockHash)
	blockNumKey := blockNumDbPrefix(genesisBlock.Number)
	return bc.db.Update(func(txn *badger.Txn) error {
		if err := txn.Set(genesisKey, genSerialized); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		if err := txn.Set(genesisBlockKey, serializedBLock); err != nil {
			bc.logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		if err := txn.Set(blockNumKey, blockKey); err != nil {
			bc.logger.Errorf("Unable to save genesis block by number: %s", err)
			return err
		}

		if err := txn.Set(blockKey, serializedBLock); err != nil {
			bc.logger.Errorf("Unable to put genesis block: %s", err)
			return err
		}

		if err := txn.Set(lastHashKey, genesisBlock.BlockHeader.BlockHash.Bytes()); err != nil {
			bc.logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		for addr := range gen.Alloc {
			alloc := gen.Alloc[addr]

			bal := types.Balance{
				Address: addr,
				Balance: alloc.Balance,
				Nonce:   0,
			}

			balanceEncoded, err := bal.Serialize()
			if err != nil {
				return err
			}

			balanceKey := balanceDbPrefix(bal.Address)
			if err := txn.Set(balanceKey, balanceEncoded); err != nil {
				bc.logger.Errorf("Unable to save balance: %s", err)
				return err
			}
		}

		bc.genesis = gen
		bc.LastHash = genesisBlock.BlockHeader.BlockHash
		return nil
	})
}

func (bc *BlockChain) loadGenesis(ctx context.Context) error {
	return bc.db.View(func(txn *badger.Txn) error {
		gen, err := txn.Get(genesisKey)
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

func (bc *BlockChain) GetGenesisBlock(ctx context.Context) (*types.Block, error) {
	if bc.tracer != nil {
		span := bc.tracer.StartSpan("get_genesis_block", traceutil.ProvideParentSpan(ctx))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	var block types.Block
	if err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(genesisBlockKey)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return block.Deserialize(val)
		})
	}); err != nil {
		if err == badger.ErrKeyNotFound {
			return nil, ErrBlockNotExists
		}
		return nil, err
	}
	return &block, nil
}
