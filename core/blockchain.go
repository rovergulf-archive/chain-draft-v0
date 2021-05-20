package core

import (
	"encoding/hex"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/spf13/viper"
	bolt "go.etcd.io/bbolt"
	"go.uber.org/zap"
	"log"
	"os"
	"path"
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
	DbFilePath string `json:"db_file_path" yaml:"db_file_path"`
	Address    string `json:"address" yaml:"address"`
	Logger     *zap.SugaredLogger
}

type Blockchain struct {
	Tip    []byte
	Db     *bolt.DB `json:"-" yaml:"-"`
	logger *zap.SugaredLogger
	tracer opentracing.Tracer
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain(opts Options) (*Blockchain, error) {
	if !dbExists() {
		return nil, fmt.Errorf("no existing blockchain [%s] found", opts.Address)
	}

	var tip []byte
	db, err := bolt.Open(dbFile(), 0600, nil)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	}); err != nil {
		return nil, err
	}

	bc := Blockchain{
		Tip:    tip,
		Db:     db,
		logger: opts.Logger,
	}

	return &bc, nil
}

// CreateBlockchain creates a new blockchain DB
func CreateBlockchain(opts Options) (*Blockchain, error) {
	if dbExists() {
		return nil, fmt.Errorf("blockchain [%s] already exists", opts.Address)
	}

	var tip []byte
	db, err := bolt.Open(dbFile(), 0600, nil)
	if err != nil {
		opts.Logger.Errorf("Unable to open db file: %s", err)
		return nil, err
	}

	if err := db.Update(func(tx *bolt.Tx) error {
		coinbase := NewCoinbaseTX(opts.Address, genesisCoinbaseData)
		genesis := NewGenesisBlock(coinbase)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			opts.Logger.Errorf("Unable to create db bucket: %s", err)
			log.Panic(err)
		}

		genSerialized, err := genesis.Serialize()
		if err != nil {
			opts.Logger.Errorf("Unable to serialize genesis block: %s", err)
			return err
		}

		if err := b.Put(genesis.Hash, genSerialized); err != nil {
			opts.Logger.Errorf("Unable to put genesis block: %s", err)
			return err
		}

		if err := b.Put([]byte("l"), genesis.Hash); err != nil {
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
		Tip:    tip,
		Db:     db,
		logger: opts.Logger,
	}

	return &bc, nil
}

// MineBlock mines a new block with the provided transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) error {
	var lastHash []byte

	if err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	}); err != nil {
		bc.logger.Errorf("Unable to get last hash: %s", err)
		return err
	}

	newBlock := NewBlock(transactions, lastHash)

	if err := bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		nb, err := newBlock.Serialize()
		if err != nil {
			bc.logger.Errorf("Unable to serialize block: %s", err)
			return err
		}

		if err := b.Put(newBlock.Hash, nb); err != nil {
			bc.logger.Errorf("Unable to add new block: %s", err)
			return err
		}

		if err := b.Put([]byte("l"), newBlock.Hash); err != nil {
			bc.logger.Errorf("Unable to update last hash: %s", err)
			return err
		}
		bc.Tip = newBlock.Hash

		return nil
	}); err != nil {
		bc.logger.Errorf("Unable to start transaction: %s", err)
		return err
	}

	return nil
}

// FindUnspentTransactions returns a list of transactions containing unspent outputs
func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block, err := bci.Next()
		if err != nil {
			break
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Was the output spent?
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO finds and returns all unspent transaction outputs
func (bc *Blockchain) FindUTXO(address string) []TXOutput {
	var UTXOs []TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// BlockchainIterator is used to iterate over blockchain blocks
type BlockchainIterator struct {
	CurrentHash []byte
	Db          *bolt.DB
	logger      *zap.SugaredLogger
}

// Iterator returns a BlockchainIterator
func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{
		CurrentHash: bc.Tip,
		Db:          bc.Db,
		logger:      bc.logger,
	}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() (*Block, error) {
	var block *Block

	if err := i.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.CurrentHash)
		nextBlock, err := DeserializeBlock(encodedBlock)
		if err != nil {
			return err
		} else {
			block = nextBlock
		}

		return nil
	}); err != nil {
		i.logger.Errorw("Unable to iterate db view",
			"current_hash", i.CurrentHash, "err", err,
		)
		return nil, err
	}

	i.CurrentHash = block.PrevHash

	return block, nil
}
