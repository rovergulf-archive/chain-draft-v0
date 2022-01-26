package core

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/chain/core/types"
	"github.com/rovergulf/chain/params"
)

func (bc *BlockChain) ApplyTx(txHash common.Hash, tx *types.SignedTx) (*types.Receipt, error) {
	fromAddr, err := bc.GetBalance(tx.From)
	if err != nil {
		bc.logger.Errorf("Unable to get sender balance: %s", err)
		return nil, err
	}

	toAddr, err := bc.GetBalance(tx.To)
	if err != nil {
		bc.logger.Errorf("Unable to get recipient balance: %s", err)
		return nil, err
	}

	if tx.Cost() > fromAddr.Balance {
		return nil, fmt.Errorf("wrong TX. Sender '%s' balance is %d TBB. Tx cost is %d TBB",
			tx.From.String(), fromAddr.Balance, tx.Cost())
	}

	fromAddr.Balance -= tx.Cost()
	toAddr.Balance += tx.Value

	fromAddr.Nonce = tx.Nonce
	toAddr.Nonce++

	from, err := fromAddr.Serialize()
	if err != nil {
		return nil, err
	}

	to, err := toAddr.Serialize()
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{
		Addr:        fromAddr.Address,
		Balance:     fromAddr.Balance,
		NetherUsed:  tx.Nether,
		NetherPrice: tx.NetherPrice,
		TxHash:      txHash,
	}

	return receipt, bc.db.Update(func(txn *badger.Txn) error {
		senderKey := append(balancesPrefix, fromAddr.Address.Bytes()...)
		if err := txn.Set(senderKey, from); err != nil {
			return err
		}

		recipientKey := append(balancesPrefix, toAddr.Address.Bytes()...)
		return txn.Set(recipientKey, to)
	})
}

func (bc *BlockChain) applyRewardTx(ctx context.Context, tx *types.SignedTx) (*types.Receipt, error) {
	if !tx.IsReward() {
		return nil, ErrInvalidRewardData
	}

	toAddr, err := bc.GetBalance(tx.To)
	if err != nil {
		bc.logger.Errorf("Unable to get recipient balance: %s", err)
		return nil, err
	}

	toAddr.Balance += tx.Value
	toAddr.Nonce++

	to, err := toAddr.Serialize()
	if err != nil {
		return nil, err
	}

	receipt := &types.Receipt{
		Addr:        tx.To,
		Balance:     toAddr.Balance,
		NetherUsed:  tx.Nether,
		NetherPrice: tx.NetherPrice,
	}

	return receipt, bc.db.Update(func(txn *badger.Txn) error {
		recipientKey := append(balancesPrefix, toAddr.Address.Bytes()...)
		return txn.Set(recipientKey, to)
	})
}

func (bc *BlockChain) ApplyBlock(ctx context.Context, block *types.Block) error {

	pool := block.NetherUsed
	bc.logger.Debugf("Nether pool available: ~%.5f", float64(pool/params.Raftel))

	var txsHashes, receiptsHashes [][]byte
	for i := range block.Transactions {
		tx := block.Transactions[i]

		hashValue, err := tx.Hash()
		if err != nil {
			return err
		}

		txsHashes = append(txsHashes, hashValue)
		txHash := common.BytesToHash(hashValue)

		var receipt *types.Receipt
		if tx.IsReward() {
			if receipt, err = bc.applyRewardTx(ctx, tx); err != nil {
				return err
			}
		} else {
			if receipt, err = bc.ApplyTx(txHash, tx); err != nil {
				return err
			}
		}

		receipt.BlockHash = block.BlockHeader.BlockHash
		receipt.BlockNumber = block.Number
		receipt.TxIndex = i
		receipt.TxHash = txHash

		if err := bc.SaveReceipt(ctx, receipt); err != nil {
			return err
		}

		rcptHash, err := receipt.Hash()
		if err != nil {
			return err
		}

		receiptsHashes = append(receiptsHashes, rcptHash)
	}

	return nil
}
