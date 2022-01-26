package types

import "time"

// TxByPriceAndTime implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPriceAndTime []*Transaction

func (txs TxByPriceAndTime) Len() int { return len(txs) }
func (txs TxByPriceAndTime) Less(i, j int) bool {
	// If the prices are equal, use the time the transaction was first seen for
	// deterministic sorting
	cmp := txs[i].Cost() - txs[j].Cost()
	if cmp == 0 {
		return time.Unix(txs[i].Time, 0).Before(time.Unix(txs[j].Time, 0))
	}
	return cmp > 0
}
func (txs TxByPriceAndTime) Swap(i, j int) { txs[i], txs[j] = txs[j], txs[i] }

func (txs *TxByPriceAndTime) Push(x interface{}) {
	*txs = append(*txs, x.(*Transaction))
}

func (txs *TxByPriceAndTime) Pop() interface{} {
	old := *txs
	n := len(old)
	x := old[n-1]
	*txs = old[0 : n-1]
	return x
}
