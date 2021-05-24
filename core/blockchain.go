package core

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DbFileName          = "chain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "Rovergulf Blockchain Genesis"
)

func dbFile() string {
	return path.Join(viper.GetString("data_dir"), DbFileName)
}

func dbExists() bool {
	if _, err := os.Stat(dbFile()); os.IsNotExist(err) {
		return false
	}

	return true
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase}, []byte{})
}

type Options struct {
	DbFilePath string         `json:"db_file_path" yaml:"db_file_path"`
	Address    string         `json:"address" yaml:"address"`
	NodeId     string         `json:"node_id" yaml:"node_id"`
	Badger     badger.Options `json:"badger" yaml:"badger"`
	Logger     *zap.SugaredLogger
	Tracer     opentracing.Tracer
}

type Blockchain struct {
	LastHash []byte
	Db       *badger.DB `json:"-" yaml:"-"`
	logger   *zap.SugaredLogger
	tracer   opentracing.Tracer
}

// ContinueBlockchain continues from existing database Blockchain
func ContinueBlockchain(opts Options) (*Blockchain, error) {
	if !dbExists() {
		return nil, fmt.Errorf("no existing blockchain [%s] found", opts.Address)
	}

	var tip []byte
	db, err := openDB(dbFile(), opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	if err := db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			tip = val
			return nil
		})
	}); err != nil {
		return nil, err
	}

	bc := Blockchain{
		LastHash: tip,
		Db:       db,
		logger:   opts.Logger,
		tracer:   opts.Tracer,
	}

	return &bc, nil
}

// InitBlockchain creates a new blockchain DB
func InitBlockchain(opts Options) (*Blockchain, error) {
	if dbExists() {
		return nil, fmt.Errorf("blockchain [%s] already exists", opts.Address)
	}

	var tip []byte

	db, err := openDB(dbFile(), opts.Badger)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	if err := db.Update(func(txn *badger.Txn) error {
		coinbase := CoinbaseTx(opts.Address, genesisCoinbaseData)
		genesis := NewGenesisBlock(coinbase)

		genSerialized, err := genesis.Serialize()
		if err != nil {
			opts.Logger.Errorf("Unable to serialize genesis block: %s", err)
			return err
		}

		if err := txn.Set(genesis.Hash, genSerialized); err != nil {
			opts.Logger.Errorf("Unable to put genesis block: %s", err)
			return err
		}

		if err := txn.Set([]byte("l"), genesis.Hash); err != nil {
			opts.Logger.Errorf("Unable to put genesis block hash: %s", err)
			return err
		}

		tip = genesis.Hash

		return nil
	}); err != nil {
		opts.Logger.Errorf("Unable to write transaction: %s", err)
		return nil, err
	}

	bc := Blockchain{
		LastHash: tip,
		Db:       db,
		logger:   opts.Logger,
		tracer:   opts.Tracer,
	}

	return &bc, nil
}

// AddBlock adds a block with the provided transactions
func (bc *Blockchain) AddBlock(block *Block) error {
	return bc.Db.Update(func(txn *badger.Txn) error {
		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}

		blockData, err := block.Serialize()
		if err != nil {
			return err
		}

		if err := txn.Set(block.Hash, blockData); err != nil {
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
					return err
				}

				if block.Height > lastBlock.Height {
					if err := txn.Set([]byte("lh"), block.Hash); err != nil {
						return err
					}
					bc.LastHash = block.Hash
				}

				return nil
			})
		})
	})
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) (*Block, error) {
	var lastHash []byte

	if err := bc.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
	}); err != nil {
		bc.logger.Errorf("Unable to get last hash: %s", err)
		return nil, err
	}

	newBlock := NewBlock(transactions, lastHash)

	if err := bc.Db.Update(func(txn *badger.Txn) error {
		nb, err := newBlock.Serialize()
		if err != nil {
			bc.logger.Errorf("Unable to serialize block: %s", err)
			return err
		}

		if err := txn.Set(newBlock.Hash, nb); err != nil {
			bc.logger.Errorf("Unable to add new block: %s", err)
			return err
		}

		if err := txn.Set([]byte("l"), newBlock.Hash); err != nil {
			bc.logger.Errorf("Unable to update last hash: %s", err)
			return err
		}
		bc.LastHash = newBlock.Hash

		return nil
	}); err != nil {
		bc.logger.Errorf("Unable to start transaction: %s", err)
		return nil, err
	}

	return newBlock, nil
}

func (bc *Blockchain) GetBestHeight() (int, error) {
	var lastBlockHeight int

	if err := bc.Db.View(func(txn *badger.Txn) error {
		lastHash, err := txn.Get([]byte("lh"))
		if err != nil {
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

				lastBlockHeight = lb.Height

				return nil
			})
		})
	}); err != nil {
		return 0, err
	}

	return lastBlockHeight, nil
}

func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	err := bc.Db.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return fmt.Errorf("block is not found")
		} else {
			return item.Value(func(val []byte) error {
				itemVal, err := DeserializeBlock(val)
				if err != nil {
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

func (bc *Blockchain) GetBlockHashes() ([][]byte, error) {
	var blocks [][]byte

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

func (bc *Blockchain) FindUTXO() (map[string]TxOutputs, error) {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := bc.Iterator()

	for {
		block, err := iter.Next()
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return UTXO, nil
}

func (bc *Blockchain) FindTransaction(ID []byte) (*Transaction, error) {
	iter := bc.Iterator()

	for {
		block, err := iter.Next()
		if err != nil {
			return nil, err
		}

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return nil, fmt.Errorf("transaction does not exist")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) error {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			return err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	return tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		if err != nil {
			return err
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = *prevTX
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) Shutdown() {
	if bc.Db != nil {
		if err := bc.Db.Close(); err != nil {
			bc.logger.Errorf("Unable to close db: %s", err)
		}
	}
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				opts.Logger.Debugf("database unlocked, value log truncated")
				return db, nil
			}
			opts.Logger.Errorf("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
