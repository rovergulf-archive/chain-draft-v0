package accounts

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"golang.org/x/crypto/ripemd160"
	"log"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	//Address    []byte           `json:"address" yaml:"address"`
	PrivateKey ecdsa.PrivateKey `json:"-" yaml:"-"`
	PublicKey  []byte           `json:"-" yaml:"-"`
}

func (w Wallet) Address() ([]byte, error) {
	pubHash, err := PublicKeyHash(w.PublicKey)
	if err != nil {
		return nil, err
	}

	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)

	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)

	return address, nil
}

func (w Wallet) StringAddr() (string, error) {
	address, err := w.Address()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s", address), nil
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

func DeserializeWallet(data []byte) (*Wallet, error) {
	var w Wallet

	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&w); err != nil {
		return nil, err
	}

	return &w, nil
}

func NewKeyPair() (*ecdsa.PrivateKey, []byte, error) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return private, pub, nil
}

func MakeWallet() (*Wallet, error) {
	private, public, err := NewKeyPair()
	if err != nil {
		return nil, err
	}

	wallet := Wallet{*private, public}

	return &wallet, nil
}

func PublicKeyHash(pubKey []byte) ([]byte, error) {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	if _, err := hasher.Write(pubHash[:]); err != nil {
		return nil, err
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD, nil
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func ValidateAddress(address string) bool {
	pubKeyHash, err := Base58Decode([]byte(address))
	if err != nil {
		return false
	}

	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
