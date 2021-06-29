package wallets

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/rovergulf/rbn/core"
	"log"
)

type Wallet struct {
	Code []byte `json:"code" yaml:"code"`
	*keystore.Key
}

func (w *Wallet) Serialize() ([]byte, error) {
	buf := bytes.Buffer{}

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(w); err != nil {
		log.Panic(err)
	}

	return buf.Bytes(), nil
}

func (w *Wallet) SignTx(tx *core.Transaction) (*core.SignedTx, error) {
	rawTx, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	sig, err := Sign(rawTx, w.Key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &core.SignedTx{
		Transaction: *tx,
		Sig:         sig,
	}, nil
}

func (w *Wallet) GetPassphrase() string {
	return string(w.Code)
}

func DeserializeWallet(data []byte) (*Wallet, error) {
	var w Wallet

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

func SignTx(tx core.Transaction, privKey *ecdsa.PrivateKey) ([]byte, error) {
	rawTx, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	sig, err := Sign(rawTx, privKey)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func NewSignedTx(tx *core.Transaction, privKey *ecdsa.PrivateKey) (*core.SignedTx, error) {
	sig, err := SignTx(*tx, privKey)
	if err != nil {
		return nil, nil
	}

	return &core.SignedTx{
		Transaction: *tx,
		Sig:         sig,
	}, nil
}

func Sign(msg []byte, privKey *ecdsa.PrivateKey) (sig []byte, err error) {
	msgHash := sha256.Sum256(msg)

	return crypto.Sign(msgHash[:], privKey)
}

func Verify(msg, sig []byte) (*ecdsa.PublicKey, error) {
	msgHash := sha256.Sum256(msg)

	recoveredPubKey, err := crypto.SigToPub(msgHash[:], sig)
	if err != nil {
		return nil, fmt.Errorf("unable to verify message signature. %s", err.Error())
	}

	return recoveredPubKey, nil
}

func NewRandomKey() (*keystore.Key, error) {
	privateKeyECDSA, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	key := &keystore.Key{
		Id:         uuid.New(),
		Address:    crypto.PubkeyToAddress(privateKeyECDSA.PublicKey),
		PrivateKey: privateKeyECDSA,
	}

	return key, nil
}
