package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

const miningIntervalSeconds = 10

type PendingBlock struct {
	parent common.Hash
	number uint64
	time   int64
	miner  common.Address
	txs    []core.SignedTx
}

func NewPendingBlock(parent common.Hash, number uint64, miner common.Address, txs []core.SignedTx) PendingBlock {
	return PendingBlock{
		parent: parent,
		number: number,
		time:   time.Now().Unix(),
		miner:  miner,
		txs:    txs,
	}
}

func Mine(ctx context.Context, pb PendingBlock) (*core.Block, error) {
	if len(pb.txs) == 0 {
		return nil, fmt.Errorf("mining empty blocks is not allowed")
	}

	start := time.Now()
	attempt := 0
	block := new(core.Block)
	var hash common.Hash
	var nonce uint64

	for !core.IsBlockHashValid(hash) {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		default:
		}

		attempt++
		nonce = generateNonce()

		if attempt%1000000 == 0 || attempt == 1 {
			fmt.Printf("Mining %d Pending TXs. Attempt: %d\n", len(pb.txs), attempt)
		}

		block = core.NewBlock(pb.parent, pb.number, nonce, pb.time, pb.miner, pb.txs)
		if err := block.SetHash(); err != nil {
			return nil, fmt.Errorf("couldn't mine block. %s", err)
		}

		hash = block.Hash
	}

	fmt.Printf("\nMined new Block '%x' using PoW%s:\n", hash, Unicode("\\U1F680"))
	fmt.Printf("\tHeight: '%v'\n", block.Number)
	fmt.Printf("\tNonce: '%v'\n", block.Nonce)
	fmt.Printf("\tCreated: '%v'\n", block.Timestamp)
	fmt.Printf("\tMiner: '%v'\n", block.Miner.String())
	fmt.Printf("\tParent: '%v'\n\n", block.PrevHash.Hex())

	fmt.Printf("\tAttempt: '%v'\n", attempt)
	fmt.Printf("\tTime: %s\n\n", time.Since(start))

	return block, nil
}

func generateNonce() uint64 {
	rand.Seed(time.Now().UTC().UnixNano())

	return rand.Uint64()
}

func Unicode(s string) string {
	r, _ := strconv.ParseInt(strings.TrimPrefix(s, "\\U"), 16, 32)

	return strconv.FormatInt(r, 10)
}
