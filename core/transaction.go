package core

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/rovergulf/rbn/accounts"
	"log"
	"math/big"
	"strings"
)

const subsidy = 10

// Transaction represents a Bitcoin transaction
type Transaction struct {
	ID      []byte     `json:"-" yaml:"-"`
	Inputs  []TxInput  `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Outputs []TxOutput `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// Hash returns a hash of the transaction
func (tx *Transaction) Hash() ([]byte, error) {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	serializedCopy, err := txCopy.Serialize()
	if err != nil {
		return nil, err
	}
	hash = sha256.Sum256(serializedCopy)

	return hash[:], nil
}

func (tx *Transaction) Serialize() ([]byte, error) {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	if err := enc.Encode(tx); err != nil {
		return nil, err
	}

	return encoded.Bytes(), nil
}

func DeserializeTransaction(data []byte) (*Transaction, error) {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&transaction); err != nil {
		return nil, err
	}

	return &transaction, nil
}

// IsCoinbase checks whether the transaction is coinbase
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// SetID sets ID of a transaction
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// CoinbaseTx creates a new coinbase transaction
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		if _, err := rand.Read(randData); err != nil {
			log.Panic(err)
		}
		data = fmt.Sprintf("%x", randData)
	}

	txIn := TxInput{
		ID:        []byte{},
		Out:       -1,
		Signature: nil,
		PubKey:    []byte{},
	}

	txOut, err := NewTxOutput(100, to)
	if err != nil {
		log.Panic(err)
	}

	tx := Transaction{nil, []TxInput{txIn}, []TxOutput{*txOut}}
	tx.SetID()

	return &tx
}

// NewTransaction creates a new transaction
func NewTransaction(w *accounts.Wallet, to string, amount int, UTXO *UTXOSet) (*Transaction, error) {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash, err := accounts.PublicKeyHash(w.PublicKey)
	if err != nil {
		return nil, err
	}

	acc, validOutputs, err := UTXO.FindSpendableOutputs(pubKeyHash, amount)
	if err != nil {
		return nil, err
	}
	fmt.Println(acc, validOutputs)

	if acc < amount {
		return nil, fmt.Errorf("error: not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			return nil, err
		}

		for _, out := range outs {
			inputs = append(inputs, TxInput{
				ID:        txID,
				Out:       out,
				Signature: nil,
				PubKey:    w.PublicKey,
			})
		}
	}

	from, err := w.StringAddr()
	if err != nil {
		return nil, err
	}

	txOutput, err := NewTxOutput(amount, to)
	if err != nil {
		return nil, err
	}
	outputs = append(outputs, *txOutput)

	if acc > amount {
		minusTxOutput, err := NewTxOutput(acc-amount, from)
		if err != nil {
			return nil, err
		}

		outputs = append(outputs, *minusTxOutput)
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID, err = tx.Hash()
	if err != nil {
		return nil, err
	}

	if err := UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey); err != nil {
		return nil, err
	}

	return &tx, nil
}

// TrimmedCopy creates a trimmed copy of Transaction to be used in signing
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	for _, vin := range tx.Inputs {
		inputs = append(inputs, TxInput{
			ID:     vin.ID,
			Out:    vin.Out,
			PubKey: []byte{},
		})
	}

	for _, vout := range tx.Outputs {
		outputs = append(outputs, TxOutput{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// Sign signs each input of a Transaction
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return fmt.Errorf("coinbase cannot be signed")
	}

	for _, vin := range tx.Inputs {
		if prevTXs[hex.EncodeToString(vin.ID)].ID == nil {
			return fmt.Errorf("ERROR: Previous transaction is not correct")
		}
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Inputs {
		txId, err := txCopy.Hash()
		if err != nil {
			return err
		}

		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[in.Out].PubKeyHash
		txCopy.ID = txId
		txCopy.Inputs[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inID].Signature = signature
	}

	return nil
}

// Verify verifies signatures of Transaction inputs
func (tx *Transaction) Verify(prevTXs map[string]Transaction) error {
	if tx.IsCoinbase() {
		return nil
	}

	for _, in := range tx.Inputs {
		if prevTXs[hex.EncodeToString(in.ID)].ID == nil {
			return fmt.Errorf("Previous transaction not correct")
		}
	}

	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.Inputs {
		txId, err := txCopy.Hash()
		if err != nil {
			return err
		}

		prevTx := prevTXs[hex.EncodeToString(vin.ID)]
		txCopy.Inputs[inID].Signature = nil
		txCopy.Inputs[inID].PubKey = prevTx.Outputs[vin.Out].PubKeyHash
		txCopy.ID = txId
		txCopy.Inputs[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return fmt.Errorf("invalid transaction")
		}
	}

	return nil
}

// String returns a human-readable representation of a transaction
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction [%x]:", tx.ID))

	for i, input := range tx.Inputs {

		lines = append(lines, fmt.Sprintf("	Input: %d:", i))
		lines = append(lines, fmt.Sprintf("	  TXID:      %x", input.ID))
		lines = append(lines, fmt.Sprintf("	  Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("	  Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("	  PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("	Output: %d:", i))
		lines = append(lines, fmt.Sprintf("	  Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("	  Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
