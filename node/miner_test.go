package node

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/wallets"
	"testing"
	"time"
)

const (
	testKsOldLadyAccount = "0x6fdc0d8d15ae6b4ebf45c52fd2aafbcbb19a65c8"
)

func TestValidBlockHash(t *testing.T) {
	hexHash := "000000fa04f8160395c387277f8b2f14837603383d33809a4db586086168edfa"
	var hash = common.Hash{}

	hex.Decode(hash[:], []byte(hexHash))

	isValid := IsBlockHashValid(hash)
	if !isValid {
		t.Fatalf("hash '%s' starting with 6 zeroes is suppose to be valid", hexHash)
	}
}

func TestInvalidBlockHash(t *testing.T) {
	hexHash := "000001fa04f8160395c387277f8b2f14837603383d33809a4db586086168edfa"
	var hash = common.Hash{}

	hex.Decode(hash[:], []byte(hexHash))

	isValid := IsBlockHashValid(hash)
	if isValid {
		t.Fatal("hash is not suppose to be valid")
	}
}

func TestMine(t *testing.T) {
	minerPrivKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivKey, miner)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	minedBlock, err := Mine(ctx, pendingBlock)
	if err != nil {
		t.Fatal(err)
	}

	if err := minedBlock.SetHash(); err != nil {
		t.Fatal(err)
	}

	if !IsBlockHashValid(minedBlock.Hash) {
		t.Fatal()
	}

	if minedBlock.Miner.String() != miner.String() {
		t.Fatal("mined block miner should equal miner from pending block")
	}
}

func TestMineWithTimeout(t *testing.T) {
	minerPrivKey, _, miner, err := generateKey()
	if err != nil {
		t.Fatal(err)
	}

	pendingBlock, err := createRandomPendingBlock(minerPrivKey, miner)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Microsecond*100)
	defer cancel()

	if _, err := Mine(ctx, pendingBlock); err == nil {
		t.Fatal(err)
	}
}

func generateKey() (*ecdsa.PrivateKey, ecdsa.PublicKey, common.Address, error) {
	privKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, ecdsa.PublicKey{}, common.Address{}, err
	}

	pubKey := privKey.PublicKey
	pubKeyBytes := elliptic.Marshal(crypto.S256(), pubKey.X, pubKey.Y)
	pubKeyBytesHash := crypto.Keccak256(pubKeyBytes[1:])

	account := common.BytesToAddress(pubKeyBytesHash[12:])

	return privKey, pubKey, account, nil
}

func createRandomPendingBlock(privKey *ecdsa.PrivateKey, acc common.Address) (PendingBlock, error) {
	tx, err := core.NewTransaction(acc, common.HexToAddress(testKsOldLadyAccount), 1, 1, []byte{})
	if err != nil {
		return PendingBlock{}, err
	}

	signedTx, err := wallets.SignTx(*tx, privKey)
	if err != nil {
		return PendingBlock{}, err
	}

	return NewPendingBlock(
		common.Hash{},
		0,
		acc,
		[]*core.SignedTx{{Transaction: *tx, Sig: signedTx}},
	), nil
}
