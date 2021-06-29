package core

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"io"
	"math/big"
	"os"
)

const (
	DbFileName     = "chain.db"
	LastHashKey    = "lh"
	ChainLengthKey = "cl"
	GenesisKey     = "g"
)

func dbExists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

type Blockchain struct {
	Balances    map[common.Address]Balance
	LastHash    common.Hash
	ChainLength *big.Int
	Db          *badger.DB `json:"-" yaml:"-"`

	genesis *Genesis

	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
}

// InitBlockchain creates a new blockchain DB
func InitBlockchain(opts config.Options) (*Blockchain, error) {
	if dbExists(opts.DbFilePath) {
		return nil, fmt.Errorf("genesis already initalized")
	}

	gen, err := loadGenesisFromFile(viper.GetString("genesis"))
	if err != nil {
		opts.Logger.Errorf("Unable to load genesis file: %s", err)
		return nil, err
	}

	genSerialized, err := gen.MarshalJSON()
	if err != nil {
		opts.Logger.Errorf("Unable to marshal genesis: %s", err)
		return nil, err
	}

	genesis, err := NewGenesisBlock(gen)
	if err != nil {
		return nil, err
	}

	genBlock, err := genesis.Serialize()
	if err != nil {
		opts.Logger.Errorf("Unable to serialize genesis block: %s", err)
		return nil, err
	}

	opts.Badger = badger.DefaultOptions(opts.DbFilePath)
	db, err := badgerdb.OpenDB(opts.DbFilePath, opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	bc := Blockchain{
		LastHash:    genesis.Hash,
		Balances:    make(map[common.Address]Balance),
		ChainLength: big.NewInt(0),
		Db:          db,
		genesis:     gen,
		logger:      opts.Logger,
		tracer:      opts.Tracer,
	}

	for addr := range gen.Alloc {
		bc.Balances[addr] = Balance{
			Address: addr,
			Balance: big.NewInt(gen.Alloc[addr].Balance),
			Nonce:   gen.Alloc[addr].Nonce,
		}
	}

	if err := db.Update(func(txn *badger.Txn) error {
		if err := txn.Set([]byte("g"), genSerialized); err != nil {
			opts.Logger.Errorf("Unable to save genesis value: %s", err)
			return err
		}

		if err := txn.Set(genesis.Hash.Bytes(), genBlock); err != nil {
			opts.Logger.Errorf("Unable to put genesis block: %s", err)
			return err
		}

		if err := txn.Set([]byte("lh"), genesis.Hash.Bytes()); err != nil {
			opts.Logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		if err := txn.Set([]byte("cl"), []byte{1}); err != nil {
			opts.Logger.Errorf("Unable to set chain length: %s", err)
			return err
		}

		for addr := range bc.Balances {
			bal := bc.Balances[addr]
			balanceEncoded, err := bal.Serialize()
			if err != nil {
				return err
			}
			fmt.Println("save balance for ", addr.Hex())
			balanceKey := append(balancesPrefix, addr.Bytes()...)
			if err := txn.Set(balanceKey, balanceEncoded); err != nil {
				opts.Logger.Errorf("Unable to save balance: %s", err)
				return err
			}
		}

		bc.LastHash = genesis.Hash

		return nil
	}); err != nil {
		opts.Logger.Errorf("Unable to write transaction: %s", err)
		return nil, err
	}

	return &bc, nil
}

// ContinueBlockchain continues from existing database Blockchain
func ContinueBlockchain(opts config.Options) (*Blockchain, error) {
	if !dbExists(opts.DbFilePath) {
		return nil, fmt.Errorf("chain db does not exists")
	}

	b := Blockchain{
		Balances:    make(map[common.Address]Balance),
		genesis:     new(Genesis),
		ChainLength: big.NewInt(0),
		logger:      opts.Logger,
		tracer:      opts.Tracer,
		closer:      opts.Closer,
	}

	opts.Badger = badger.DefaultOptions(opts.DbFilePath)
	db, err := badgerdb.OpenDB(opts.DbFilePath, opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	if err := db.View(func(txn *badger.Txn) error {
		lh, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return lh.Value(func(val []byte) error {
			b.LastHash = common.BytesToHash(val)

			chainLength, err := txn.Get([]byte("cl"))
			if err != nil {
				return err
			}

			return chainLength.Value(func(val []byte) error {
				b.ChainLength.SetBytes(val)
				return nil
			})
		})
	}); err != nil {
		return nil, err
	}

	b.Db = db
	return &b, nil
}

// AddBlock adds a block with the provided transactions
func (bc *Blockchain) AddBlock(block *Block) error {
	if len(block.Hash.Bytes()) == 0 {
		return fmt.Errorf("bad block hash")
	}

	return bc.Db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash.Bytes()); err == nil {
			return nil
		}

		blockData, err := block.Serialize()
		if err != nil {
			return err
		}

		if err := txn.Set(block.Hash.Bytes(), blockData); err != nil {
			return err
		}

		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			item, err := txn.Get(val)
			if err != nil {
				return err
			}

			return item.Value(func(val []byte) error {
				lastBlock, err := DeserializeBlock(val)
				if err != nil {
					bc.logger.Errorf("Unable to deserialize block: %s", err)
					return err
				}

				if block.Number > lastBlock.Number {
					if err := txn.Set([]byte("lh"), block.Hash.Bytes()); err != nil {
						bc.logger.Errorf("Unable to set last hash value: %s", err)
						return err
					}
					bc.LastHash = block.Hash
				}

				bc.logger.Infow("Saved block", "prev", lastBlock.Hash,
					"hash", block.Hash, "number", block.Number, "txs", len(block.Transactions))

				return nil
			})
		})
	})
}

func (bc *Blockchain) GetBestHeight() (uint64, error) {
	var lastBlockHeight uint64

	if err := bc.Db.View(func(txn *badger.Txn) error {
		lastHash, err := txn.Get([]byte("lh"))
		if err != nil {
			bc.logger.Errorf("Unable to get last hash value: %s", err)
			return err
		}

		return lastHash.Value(func(val []byte) error {
			lastBlockData, err := txn.Get(val)
			if err != nil {
				return err
			}

			return lastBlockData.Value(func(val []byte) error {
				lb, err := DeserializeBlock(val)
				if err != nil {
					return err
				}

				lastBlockHeight = lb.Number

				return nil
			})
		})
	}); err != nil {
		return 0, err
	}

	return lastBlockHeight, nil
}

func (bc *Blockchain) GetBlock(hash common.Hash) (Block, error) {
	var block Block

	err := bc.Db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(hash.Bytes()); err != nil {
			return fmt.Errorf("block is not found")
		} else {
			return item.Value(func(val []byte) error {
				itemVal, err := DeserializeBlock(val)
				if err != nil {
					bc.logger.Errorf("Unable to deserialize block: %s", err)
					return err
				} else {
					block = *itemVal
				}
				return nil
			})
		}
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (bc *Blockchain) GetBlockHashes() ([]common.Hash, error) {
	var blocks []common.Hash

	iter := bc.Iterator()

	for {
		block, err := iter.Next()
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks, nil
}

func (bc *Blockchain) GetGenesis() (*Genesis, error) {
	var g Genesis
	if err := bc.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("g"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return g.UnmarshalJSON(val)
		})
	}); err != nil {
		return nil, err
	}

	return &g, nil
}

func (bc *Blockchain) FindTransaction(txId []byte) (*SignedTx, error) {
	var tx SignedTx

	if err := bc.Db.View(func(txn *badger.Txn) error {
		key := append(txPrefix, txId...)
		item, err := txn.Get(key)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return tx.Deserialize(val)
		})
	}); err != nil {
		return nil, err
	}

	return &tx, nil
}

func (bc *Blockchain) Shutdown() {
	if bc.Db != nil {
		if err := bc.Db.Close(); err != nil {
			bc.logger.Errorf("Unable to close db: %s", err)
		}
	}
}
