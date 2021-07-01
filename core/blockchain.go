package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
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

var (
	ErrChainNotExists = fmt.Errorf("chain db does not exists")
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

	db     *badger.DB
	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
}

func (bc *Blockchain) Shutdown() {
	if bc.db != nil {
		if err := bc.db.Close(); err != nil {
			bc.logger.Errorf("Unable to close db: %s", err)
		}
	}
}

// InitBlockchain creates a new blockchain DB
func InitBlockchain(opts params.Options) (*Blockchain, error) {
	if dbExists(opts.DbFilePath) {
		return nil, fmt.Errorf("genesis already initalized")
	}

	gen, err := loadGenesisFromFile(viper.GetString("genesis"))
	if err != nil {
		opts.Logger.Errorf("Unable to load genesis file: %s", err)
		return nil, err
	}

	if err := gen.Validate(); err != nil {
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
		LastHash: common.Hash{},
		Balances: make(map[common.Address]Balance),
		db:       db,
		logger:   opts.Logger,
		tracer:   opts.Tracer,
	}

	for addr := range gen.Alloc {
		bc.Balances[addr] = Balance{
			Address: addr,
			Balance: gen.Alloc[addr].Balance,
			Nonce:   gen.Alloc[addr].Nonce,
		}
	}

	bc.Balances[gen.Coinbase] = Balance{
		Address: gen.Coinbase,
		Balance: 0,
		Nonce:   1,
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

		if err := txn.Set([]byte("rh"), genesis.Hash.Bytes()); err != nil {
			opts.Logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		if err := txn.Set([]byte("lh"), genesis.Hash.Bytes()); err != nil {
			opts.Logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		for addr := range bc.Balances {
			bal := bc.Balances[addr]
			balanceEncoded, err := bal.Serialize()
			if err != nil {
				return err
			}

			balanceKey := append(balancesPrefix, addr.Bytes()...)
			if err := txn.Set(balanceKey, balanceEncoded); err != nil {
				opts.Logger.Errorf("Unable to save balance: %s", err)
				return err
			}
		}

		return nil
	}); err != nil {
		opts.Logger.Errorf("Unable to write transaction: %s", err)
		return nil, err
	}

	return &bc, nil
}

// ContinueBlockchain continues from existing database Blockchain
func ContinueBlockchain(opts params.Options) (*Blockchain, error) {
	b := Blockchain{
		Balances:    make(map[common.Address]Balance),
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
			// is it ok??
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return lh.Value(func(val []byte) error {
			b.LastHash = common.BytesToHash(val)

			return nil
		})
	}); err != nil {
		return nil, err
	}

	b.db = db
	return &b, nil
}

func (bc *Blockchain) DbSize() (int64, int64) {
	return bc.db.Size()
}

// AddBlock adds a block with the provided transactions
func (bc *Blockchain) AddBlock(block *Block) error {
	if len(block.Hash.Bytes()) == 0 {
		return fmt.Errorf("bad block hash")
	}

	blockData, err := block.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash.Bytes()); err == nil {
			return nil
		}

		lastHash, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return lastHash.Value(func(val []byte) error {
			lb, err := txn.Get(val)
			if err != nil {
				return err
			}

			return lb.Value(func(val []byte) error {
				lastBlock, err := DeserializeBlock(val)
				if err != nil {
					bc.logger.Errorf("Unable to deserialize block: %s", err)
					return err
				}

				if block.Number > lastBlock.Number {
					if err := txn.Set(block.Hash.Bytes(), blockData); err != nil {
						return err
					}

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

func (bc *Blockchain) GetBlock(hash common.Hash) (Block, error) {
	var block Block

	err := bc.db.View(func(txn *badger.Txn) error {
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

		if bytes.Compare(block.PrevHash.Bytes(), common.Hash{}.Bytes()) == 0 {
			break
		}
	}

	return blocks, nil
}

func (bc *Blockchain) GetGenesis() (*Genesis, error) {
	var g Genesis
	if err := bc.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("g"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &g)
		})
	}); err != nil {
		return nil, err
	}

	return &g, nil
}

func (bc *Blockchain) FindTransaction(txId []byte) (*SignedTx, error) {
	var tx SignedTx

	if err := bc.db.View(func(txn *badger.Txn) error {
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

func (bc *Blockchain) SaveTx(tx SignedTx) error {
	encodedTx, err := tx.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		key := append(txPrefix, tx.Hash.Bytes()...)
		return txn.Set(key, encodedTx)
	})
}

func (bc *Blockchain) ApplyTx(tx SignedTx) error {
	fromAddr, err := bc.GetBalance(tx.From)
	if err != nil {
		return err
	}

	toAddr, err := bc.GetBalance(tx.To)
	if err != nil {
		return err
	}

	if tx.Cost() > fromAddr.Balance {
		return fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB",
			tx.From.String(), fromAddr.Balance, tx.Cost())
	}

	fromAddr.Balance -= tx.Cost()
	toAddr.Balance += tx.Value

	fromAddr.Nonce = tx.Nonce

	from, err := fromAddr.Serialize()
	if err != nil {
		return err
	}

	to, err := toAddr.Serialize()
	if err != nil {
		return err
	}

	return bc.db.Update(func(txn *badger.Txn) error {
		senderKey := append(balancesPrefix, fromAddr.Address.Bytes()...)
		if err := txn.Set(senderKey, from); err != nil {
			return err
		}

		recipientKey := append(balancesPrefix, toAddr.Address.Bytes()...)
		return txn.Set(recipientKey, to)
	})
}

func IsBlockHashValid(hash common.Hash) bool {
	return fmt.Sprintf("%x", hash[0]) == "0" &&
		fmt.Sprintf("%x", hash[1]) == "0" &&
		fmt.Sprintf("%x", hash[2]) == "0" &&
		fmt.Sprintf("%x", hash[3]) != "0"
}
