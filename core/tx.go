package core

import (
	"bytes"
	"encoding/gob"
	"github.com/rovergulf/rbn/accounts"
)

// TxInput represents a transaction input
type TxInput struct {
	ID        []byte
	Out       int
	Signature []byte
	PubKey    []byte
}

// UsesKey checks whether the address initiated the transaction
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := accounts.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// TxOutput represents a transaction output
type TxOutput struct {
	Value      int
	PubKeyHash []byte
}

type TxOutputs struct {
	Outputs []TxOutput
}

// Lock signs the output
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := accounts.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTxOutput create a new TxOutput
func NewTxOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

func (outs TxOutputs) Serialize() ([]byte, error) {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	if err := encode.Encode(outs); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func DeserializeOutputs(data []byte) (TxOutputs, error) {
	var outputs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(data))
	if err := decode.Decode(&outputs); err != nil {
		return outputs, err
	}

	return outputs, nil
}
