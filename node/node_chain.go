package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core/types"
	"time"
)

func (n *Node) generateBlock(ctx context.Context) error {
	if len(n.pendingTXs) == 0 {
		return fmt.Errorf("no transactions available")
	}

	var txs []types.SignedTx

	for i := range n.pendingTXs {
		tx := n.pendingTXs[i]
		txs = append(txs, tx)
	}

	lb, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		return err
	}

	header := types.BlockHeader{
		PrevHash:  lb.BlockHeader.Hash,
		Number:    lb.Number + 1,
		Timestamp: time.Now().Unix(),
		Coinbase:  n.account.Address(),
	}

	b := types.NewBlock(header, txs)
	hash, err := b.Hash()
	if err != nil {
		return err
	}
	b.BlockHeader.Hash = common.BytesToHash(hash)

	var receipts []*types.Receipt
	var txHashes, receiptsHashes [][]byte
	for _, tx := range b.Transactions {
		hash, err := tx.Hash()
		if err != nil {
			return err
		}
		txHashes = append(txHashes, hash)
		b.NetherUsed += tx.Nether
	}
	b.TxHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	for _, rcp := range receipts {
		hash, err := rcp.Hash()
		if err != nil {
			return err
		}
		receiptsHashes = append(receiptsHashes, hash)
	}
	b.ReceiptHash = sha256.Sum256(bytes.Join(receiptsHashes, []byte{}))

	if err := n.bc.AddBlock(b); err != nil {
		return err
	}

	n.removeAppliedPendingTXs(b)

	return nil
}

func (n *Node) removeAppliedPendingTXs(block *types.Block) {
	if len(block.Transactions) > 0 && len(n.pendingTXs) > 0 {
		n.logger.Info("Updating in-memory Pending TXs Pool:")
	}

	for _, tx := range block.Transactions {
		txHash, err := tx.Transaction.Hash()
		if err != nil {
			n.logger.Warnf("Unable to get transaction hash: %s", err)
			continue
		}

		hash := common.BytesToHash(txHash)
		if _, exists := n.pendingTXs[hash]; exists {
			n.logger.Infof("Archiving mined TX: %s", hash)

			delete(n.pendingTXs, hash)
		}
	}
}

func (n *Node) AddPendingTX(tx types.SignedTx, peer PeerNode) (*types.Receipt, error) {
	ok, err := tx.IsAuthentic()
	if err != nil {
		return nil, err
	}

	if !ok {
		// TODO set report counter and attacker account purge
		return nil, fmt.Errorf("wrong TX. Sender '%s' is forged", tx.From)
	}

	txHash, err := tx.Transaction.Hash()
	if err != nil {
		return nil, err
	}

	hash := common.BytesToHash(txHash)

	if err := n.bc.SaveTx(hash, tx); err != nil {
		return nil, err
	}

	receipt, err := n.bc.ApplyTx(hash, tx)
	if err != nil {
		return nil, err
	}

	n.pendingTXs[hash] = tx

	return receipt, nil
}
