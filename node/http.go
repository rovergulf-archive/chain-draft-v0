package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/wallets"
	"net/http"
)

func (n *Node) serveHttp() error {
	r := n.httpHandler.router

	r.HandleFunc("/health", n.healthCheck).Methods(http.MethodGet)
	r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
	r.HandleFunc("/metrics/json", n.DiscoverMetrics).Methods(http.MethodGet)
	r.HandleFunc("/routes", n.WalkRoutes).Methods(http.MethodGet)

	r.HandleFunc(endpointStatus, n.healthCheck).Methods(http.MethodGet)
	r.HandleFunc(endpointAddPeer, n.AddPeerNode).Methods(http.MethodGet)
	r.HandleFunc(endpointSync, n.SyncPeers).Methods(http.MethodGet)

	r.HandleFunc("/node/info", n.nodeInfo).Methods(http.MethodGet)
	r.HandleFunc("/chain/info", n.healthCheck).Methods(http.MethodGet)

	r.HandleFunc("/genesis", n.ShowGenesis).Methods(http.MethodGet)
	r.HandleFunc("/blocks", n.ListBlocks).Methods(http.MethodGet)
	r.HandleFunc("/blocks/latest", n.LatestBlock).Methods(http.MethodGet)
	r.HandleFunc("/block/{hash}", n.FindBlock).Methods(http.MethodGet)
	r.HandleFunc("/balances", n.ListBalances).Methods(http.MethodGet)
	r.HandleFunc("/balances/{addr}", n.GetBalance).Methods(http.MethodGet)
	r.HandleFunc("/tx/add", n.txAdd).Methods(http.MethodPost)
	r.HandleFunc("/tx/{hash}", n.txFind).Methods(http.MethodGet)

	r.HandleFunc("/accounts", n.healthCheck).Methods(http.MethodGet)
	r.HandleFunc("/accounts", n.healthCheck).Methods(http.MethodPost)
	r.HandleFunc("/accounts", n.healthCheck).Methods(http.MethodPut)
	r.HandleFunc("/accounts/{address}", n.healthCheck).Methods(http.MethodGet)

	return http.ListenAndServe(n.metadata.ApiAddress(), &n.httpHandler)
}

func (n *Node) nodeInfo(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	lb, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			n.httpResponse(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	bcLsm, bcVlog := n.bc.DbSize() // chain db size
	wLsm, wVlog := n.wm.DbSize()   // wallets db size
	nLsm, nVlog := n.db.Size()     // node db size

	n.httpResponse(w, map[string]interface{}{
		"lash_hash":   lb.BlockHeader.Hash.Hex(),
		"pending_txs": len(n.pendingTXs),
		"in_gen_race": n.inGenRace,
		"db_size": map[string]int64{
			"chain_lsm":    bcLsm,
			"chain_vlog":   bcVlog,
			"wallets_lsm":  wLsm,
			"wallets_vlog": wVlog,
			"node_lsm":     nLsm,
			"node_vlog":    nVlog,
		},
	})
}

func (n *Node) ShowGenesis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	gen, err := n.bc.GetGenesis(ctx)
	if err != nil {
		n.httpResponse(w, true, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, gen)
}

func (n *Node) AddPeerNode(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	n.logger.Debug("http server AddPeerNode called")
	n.httpResponse(w, true, http.StatusNotImplemented)
}

func (n *Node) SyncPeers(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	n.logger.Debug("http server SyncPeers called")
	n.httpResponse(w, true, http.StatusNotImplemented)
}

func (n *Node) ListBalances(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	balances, err := n.bc.ListBalances()
	if err != nil {
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, balances)
}

func (n *Node) GetBalance(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	vars := mux.Vars(r)
	addr := vars["addr"]

	if !common.IsHexAddress(addr) {
		n.httpResponse(w, fmt.Errorf("invalid address: %s", addr), http.StatusBadRequest)
		return
	}

	address := common.HexToAddress(addr)

	balance, err := n.bc.GetBalance(address)
	if err != nil {
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	n.httpResponse(w, balance)
}

func (n *Node) txAdd(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	var req TxAddReq
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	from := common.HexToAddress(req.From)

	if from.String() == common.HexToAddress("").String() {
		n.httpResponse(w, fmt.Errorf("%s is an invalid 'from' sender", from.String()))
		return
	}

	if req.FromPwd == "" {
		n.httpResponse(w, fmt.Errorf("passphrase to decrypt the '%s' account is required. 'from_pwd' is empty",
			from.String()), http.StatusBadRequest)
		return
	}

	nonce := n.bc.GetNextAccountNonce(from)
	tx, err := types.NewTransaction(from, common.HexToAddress(req.To), req.Value, nonce, req.Data)
	if err != nil {
		n.logger.Errorf("Unable to create new transaction: %s", err)
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	wallet, err := n.wm.GetWallet(from, req.FromPwd)
	if err != nil {
		n.logger.Errorf("Unable to find stored account key: %s", err)
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	signedTx, err := wallets.NewSignedTx(tx, wallet.GetKey().PrivateKey)
	if err != nil {
		n.logger.Errorf("Unable to sign tx: %s", err)
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	receipt, err := n.AddPendingTX(signedTx, n.metadata)
	if err != nil {
		n.logger.Errorf("Unable to add pending tx: %s", err)
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, receipt)
}

func (n *Node) txFind(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	vars := mux.Vars(r)
	hashVar := vars["hash"]
	if len(hashVar) == 0 {
		n.httpResponse(w, fmt.Errorf("invalid hash"), http.StatusBadRequest)
		return
	}

	hash := common.HexToHash(hashVar)
	if bytes.Compare(hash.Bytes(), common.HexToHash("").Bytes()) == 0 {
		n.httpResponse(w, fmt.Errorf("invalid hash"), http.StatusBadRequest)
		return
	}

	tx, err := n.bc.FindTransaction(hash.Bytes())
	if err != nil {
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, tx)
}

func (n *Node) ListBlocks(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	blocks, err := n.bc.GetBlockHashes()
	if err != nil {
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, blocks)
}

func (n *Node) LatestBlock(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	b, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, b)
}

func (n *Node) FindBlock(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	vars := mux.Vars(r)
	hash := vars["hash"]
	if len(hash) == 0 {
		n.httpResponse(w, fmt.Errorf("invalid hash: %s", hash), http.StatusBadRequest)
		return
	}

	b, err := n.bc.GetBlock(common.HexToHash(hash))
	if err != nil {
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, b)
}
