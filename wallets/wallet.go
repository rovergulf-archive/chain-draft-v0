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
	"github.com/rovergulf/chain/core/types"
	"github.com/tyler-smith/go-bip39"
)

func init() {
	gob.Register(elliptic.P256())
}

const (
	WalletStatusLocked   = "Locked"
	WalletStatusUnlocked = "Unlocked"
)

type Wallet struct {
	Auth    string `json:"auth" yaml:"auth"`
	KeyData []byte `json:"-" yaml:"-"` // stores encrypted key
	key     *keystore.Key
}

func (w *Wallet) Serialize() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(w); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (w *Wallet) Deserialize(data []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	return decoder.Decode(w)
}

func (w *Wallet) SignTx(tx *types.Transaction) (*types.SignedTx, error) {
	if w.key == nil {
		return nil, ErrAccountIsLocked
	}

	rawTx, err := tx.Serialize()
	if err != nil {
		return nil, err
	}

	sig, err := Sign(rawTx, w.key.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &types.SignedTx{
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

func (w *Wallet) Status() string {
	if w.key != nil {
		return WalletStatusUnlocked
	} else {
		return WalletStatusLocked
	}
}

func (w *Wallet) Open() error {
	key, err := keystore.DecryptKey(w.KeyData, w.Auth)
	if err != nil {
		return err
	}

	w.key = key
	return nil
}

func (w *Wallet) EncryptKey() error {
	data, err := keystore.EncryptKey(w.key, w.Auth, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	w.KeyData = data
	return nil
}

func SignTx(tx types.Transaction, privKey *ecdsa.PrivateKey) ([]byte, error) {
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

func NewSignedTx(tx types.Transaction, privKey *ecdsa.PrivateKey) (types.SignedTx, error) {
	sig, err := SignTx(tx, privKey)
	if err != nil {
		return types.SignedTx{}, nil
	}

	return types.SignedTx{
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

func NewRandomMnemonic() (string, error) {
	entropy, err := bip39.NewEntropy(256)
	if err != nil {
		return "", err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", err
	}
	return mnemonic, nil
}
