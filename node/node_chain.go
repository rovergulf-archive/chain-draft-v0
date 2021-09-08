package node

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/params"
	"time"
)

var (
	ErrNoTxAvailable = fmt.Errorf("no transactions available")
)

func (n *Node) generateBlock(ctx context.Context) (*types.Block, error) {
	if n.tracer != nil {
		span := n.tracer.StartSpan("generate_block")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if n.pendingState.pendingTxLen() == 0 {
		return nil, ErrNoTxAvailable
	}

	txs := n.pendingState.getTxsAsArray(params.TxPerBlockLimit)

	lb, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		return nil, err
	}

	header := types.BlockHeader{
		Root:      lb.Root,
		PrevHash:  lb.BlockHash,
		Number:    lb.Number + 1,
		Timestamp: time.Now().Unix(),
		Coinbase:  n.account.Address(),
	}

	if core.IsHashEmpty(header.Root) {
		gb, err := n.bc.GetGenesisBlock(ctx)
		if err != nil {
			return nil, err
		}
		header.Root = gb.BlockHash
	}

	b := types.NewBlock(header, txs)

	var txHashes [][]byte
	for _, tx := range b.Transactions {
		txHash, err := tx.Hash()
		if err != nil {
			return nil, err
		}
		txHashes = append(txHashes, txHash)
		b.NetherUsed += tx.Nether
	}

	rewardTxs, err := n.genRewardTxs(b)
	if err != nil {
		return nil, err
	}
	b.Transactions = append(b.Transactions, rewardTxs...)
	txs = append(txs, rewardTxs...)

	b.TxHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	blockHash, err := b.Hash()
	if err != nil {
		return nil, err
	}
	b.BlockHash = common.BytesToHash(blockHash)

	n.removeAppliedPendingTXs(ctx, b)

	if err := n.bc.AddBlock(b); err != nil {
		return nil, err
	}

	if err := n.bc.ApplyBlock(ctx, b); err != nil {
		return nil, err
	}

	return b, nil
}

func (n *Node) genRewardTxs(b *types.Block) ([]*types.SignedTx, error) {
	var txs []*types.SignedTx

	peers := n.knownPeers.GetPeers()

	// add this node to separate reward between all peers
	peers[n.metadata.TcpAddress()] = n.metadata

	peersAward := b.NetherUsed / uint64(len(peers))

	perPeer := peersAward / uint64(len(peers))

	for addr := range peers {
		peer := peers[addr]

		amount := perPeer
		if peer.account == b.Coinbase {
			amount += params.NetherLimit
		}

		tx, err := types.NewTransaction(common.HexToAddress(""), peer.account, amount, 0, types.TxRewardData)
		if err != nil {
			return nil, err
		}

		signedTx, err := n.account.SignTx(&tx)
		if err != nil {
			return nil, err
		}

		txs = append(txs, signedTx)
	}

	return txs, nil
}

func (n *Node) removeAppliedPendingTXs(ctx context.Context, block *types.Block) {
	if n.tracer != nil {
		span := n.tracer.StartSpan("add_pending_tx")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	pendingTxLen := n.pendingState.pendingTxLen()
	if len(block.Transactions) > 0 && pendingTxLen > 0 {
		n.logger.Info("Updating in-memory Pending TXs Pool:")
	}

	for _, tx := range block.Transactions {
		txHash, err := tx.Transaction.Hash()
		if err != nil {
			n.logger.Warnf("Unable to get transaction hash: %s", err)
			continue
		}

		n.pendingState.removeTx(common.BytesToHash(txHash))
	}
}

func (n *Node) AddPendingTX(ctx context.Context, tx types.SignedTx, peer PeerNode) (*types.Receipt, error) {
	if n.tracer != nil {
		span := n.tracer.StartSpan("add_pending_tx")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if tx.IsReward() {
		fmt.Println("\t-_-\treward tx")
	}

	ok, err := tx.IsAuthentic()
	if err != nil {
		return nil, err
	}

	if !ok {
		// TODO set report counter and attacker account purge
		return nil, fmt.Errorf("wrong TX. Sender '%s' is forged", tx.From)
	}

	balance, ok := n.pendingState.getBalance(tx.From)
	if !ok {
		accountBalance, err := n.bc.GetBalance(tx.From)
		if err != nil {
			return nil, err
		} else {
			if err := n.pendingState.addBalance(tx.From, accountBalance); err != nil {
				return nil, err
			}
			balance = accountBalance
		}
	}

	txHash, err := tx.Transaction.Hash()
	if err != nil {
		return nil, err
	}

	hash := common.BytesToHash(txHash)

	balance.Balance -= tx.Cost()
	balance.Nonce = tx.Nonce

	receipt := &types.Receipt{
		Addr:        tx.From,
		Status:      0,
		Balance:     balance.Balance,
		NetherUsed:  tx.Nether,
		NetherPrice: tx.NetherPrice,
		TxHash:      hash,
		TxIndex:     0,
	}

	if err := n.pendingState.addTx(hash, &tx); err != nil {
		return nil, err
	}

	return receipt, nil
}
