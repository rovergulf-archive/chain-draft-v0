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
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/rovergulf/rbn/core"
	"log"
)

func init() {
	gob.Register(elliptic.P256())
}

type Wallet struct {
	Auth string `json:"auth" yaml:"auth"`
	Data []byte `json:"-" yaml:"-"` // stores encrypted key
	key  *keystore.Key
}

func (w *Wallet) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(w); err != nil {
		log.Panic(err)
	}

	return buf.Bytes(), nil
}

func (w *Wallet) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(w)
}

func (w *Wallet) SignTx(tx *core.Transaction) (*core.SignedTx, error) {
	rawTx, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	sig, err := Sign(rawTx, w.key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &core.SignedTx{
		Transaction: *tx,
		Sig:         sig,
	}, nil
}

func (w *Wallet) Address() common.Address {
	return w.key.Address
}

func (w *Wallet) GetKey() *keystore.Key {
	return w.key
}

func (w *Wallet) Open() (*keystore.Key, error) {
	key, err := keystore.DecryptKey(w.Data, w.Auth)
	if err != nil {
		return nil, err
	}

	return key, nil
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

func NewSignedTx(tx core.Transaction, privKey *ecdsa.PrivateKey) (core.SignedTx, error) {
	sig, err := SignTx(tx, privKey)
	if err != nil {
		return core.SignedTx{}, nil
	}

	return core.SignedTx{
		Transaction: tx,
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
