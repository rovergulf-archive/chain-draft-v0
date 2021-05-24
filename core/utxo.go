package core

import (
	"bytes"
	"encoding/hex"
	"github.com/dgraph-io/badger/v3"
)

var (
	utxoPrefix   = []byte("utxo-")
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct {
	Blockchain *Blockchain
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	unspentOuts := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db

	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()

			var outs TxOutputs
			if err := item.Value(func(val []byte) error {
				outputs, err := DeserializeOutputs(val)
				if err != nil {
					return err
				} else {
					outs = outputs
				}
				return nil
			}); err != nil {
				return err
			}

			k = bytes.TrimPrefix(k, utxoPrefix)
			txID := hex.EncodeToString(k)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx)
				}
			}
		}
		return nil
	}); err != nil {
		return 0, nil, err
	}

	return accumulated, unspentOuts, nil
}

func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) ([]TxOutput, error) {
	var UTXOs []TxOutput

	db := u.Blockchain.Db

	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()

			var outs TxOutputs
			if err := item.Value(func(val []byte) error {
				outputs, err := DeserializeOutputs(val)
				if err != nil {
					return err
				} else {
					outs = outputs
				}
				return nil
			}); err != nil {
				return err
			}

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}

		}
		return nil
	}); err != nil {
		return nil, err
	}

	return UTXOs, nil
}

func (u UTXOSet) CountTransactions() (int, error) {
	db := u.Blockchain.Db
	counter := 0

	if err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	}); err != nil {
		return 0, err
	}

	return counter, nil
}

func (u UTXOSet) Reindex() error {
	db := u.Blockchain.Db

	if err := u.DeleteByPrefix(utxoPrefix); err != nil {
		return err
	}

	UTXO, err := u.Blockchain.FindUTXO()
	if err != nil {
		return err
	}

	return db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			if err != nil {
				return err
			}
			key = append(utxoPrefix, key...)

			outputs, err := outs.Serialize()
			if err != nil {
				return err
			}

			if err := txn.Set(key, outputs); err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *UTXOSet) Update(block *Block) error {
	db := u.Blockchain.Db

	return db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			// if not coinbase
			if !tx.IsCoinbase() {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					if err != nil {
						return err
					}

					var outs TxOutputs
					if err := item.Value(func(val []byte) error {
						outputs, err := DeserializeOutputs(val)
						if err != nil {
							return err
						} else {
							outs = outputs
						}
						return nil
					}); err != nil {
						return err
					}

					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							return err
						}
					} else {
						serializedOutputs, err := updatedOuts.Serialize()
						if err != nil {
							return err
						}

						if err := txn.Set(inID, serializedOutputs); err != nil {
							return err
						}
					}
				}
			}
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			serializedOutputs, err := newOutputs.Serialize()
			if err != nil {
				return err
			}

			if err := txn.Set(txID, serializedOutputs); err != nil {
				return err
			}
		}

		return nil
	})
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte) error {
	deleteKeys := func(keysForDelete [][]byte) error {
		if err := u.Blockchain.Db.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000
	return u.Blockchain.Db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				if err := deleteKeys(keysForDelete); err != nil {
					return err
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}

		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil {
				return err
			}
		}

		return nil
	})
}
