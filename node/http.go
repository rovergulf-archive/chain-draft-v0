package node

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/version"
	"github.com/rovergulf/rbn/wallets"
	"net/http"
	"strings"
	"time"
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

func (n *Node) healthCheck(w http.ResponseWriter, r *http.Request) {
	n.httpResponse(w, map[string]interface{}{
		"http_status": http.StatusOK,
		"timestamp":   time.Now().Unix(),
		"run_date":    version.RunDate.Format(time.RFC1123),
		"node_status": "healthy",
		"is_mining":   n.isMining,
	})
}

func (n *Node) DiscoverMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	metricsUrl := fmt.Sprintf("%s/metrics", n.metadata.HttpApiAddress())
	req, err := http.Get(metricsUrl)
	if err != nil {
		n.logger.Errorf("Unable to send request to prometheus metrics: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		n.httpResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	mfChan := make(chan *dto.MetricFamily, 1024)

	// Missing input means we are reading from an URL.
	if err := prom2json.ParseReader(req.Body, mfChan); err != nil {
		n.logger.Errorf("error reading metrics: %s", err)
		n.httpResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var result []*prom2json.Family
	for mf := range mfChan {
		result = append(result, prom2json.NewFamily(mf))
	}

	n.httpResponse(w, result)
}

func (n *Node) WalkRoutes(w http.ResponseWriter, r *http.Request) {
	var results []map[string]interface{}

	err := n.httpHandler.router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		res := make(map[string]interface{})

		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			//h.Logger.Debug("ROUTE: ", pathTemplate)
			res["route"] = pathTemplate
		}
		pathRegexp, err := route.GetPathRegexp()
		if err == nil {
			//h.Logger.Debug("Path regexp: ", pathRegexp)
			res["regexp"] = pathRegexp
		}
		queriesTemplates, err := route.GetQueriesTemplates()
		if err == nil {
			//h.Logger.Debug("Queries templates: ", strings.Join(queriesTemplates, ","))
			res["queries_templates"] = strings.Join(queriesTemplates, ",")
		}
		queriesRegexps, err := route.GetQueriesRegexp()
		if err == nil {
			//h.Logger.Debug("Queries regexps: ", strings.Join(queriesRegexps, ","))
			res["queries_regexps"] = strings.Join(queriesRegexps, ",")
		}
		methods, err := route.GetMethods()
		if err == nil {
			//h.Logger.Debug("Methods: ", strings.Join(methods, ","))
			res["methods"] = methods
		}

		results = append(results, res)
		return nil
	})
	if err != nil {
		n.logger.Error(err)
	}

	n.httpResponse(w, results)
}

func (n *Node) nodeInfo(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	lb, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		n.httpResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	lsm, vlog := n.bc.DbSize()
	wLsm, wVlog := n.wm.DbSize()
	nLsm, nVlog := n.db.Size()
	result := StatusRes{
		LastHash:   lb.Hash.Hex(),
		Number:     n.bc.ChainLength.Uint64(),
		KnownPeers: n.knownPeers,
		PendingTXs: n.pendingTXs,
		IsMining:   n.isMining,
		DbSize: map[string]int64{
			"chain_lsm":    lsm,
			"chain_vlog":   vlog,
			"wallets_lsm":  wLsm,
			"wallets_vlog": wVlog,
			"node_lsm":     nLsm,
			"node_vlog":    nVlog,
		},
	}

	n.httpResponse(w, result)
}

func (n *Node) ShowGenesis(w http.ResponseWriter, r *http.Request) {
	//ctx := r.Context()
	gen, err := n.bc.GetGenesis()
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
	n.logger.Debug("http server ListBalances called")

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

	balance, err := n.bc.GetBalance(common.HexToAddress(addr))
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
	tx, err := core.NewTransaction(from, common.HexToAddress(req.To), req.Value, nonce, req.Data)
	if err != nil {
		n.logger.Errorf("Unable to create new transaction: %s", err)
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	if err := tx.SetHash(); err != nil {
		n.logger.Errorf("Unable to set transaction hash: %s", err)
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	storedKey, err := n.wm.FindAccountKey(from)
	if err != nil {
		n.logger.Errorf("Unable to find stored account key: %s", err)
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	key, err := keystore.DecryptKey(storedKey, req.FromPwd)
	if err != nil {
		n.logger.Errorf("Unable to find stored account key: %s", err)
		n.httpResponse(w, err, http.StatusBadRequest)
		return
	}

	signedTx, err := wallets.NewSignedTx(tx, key.PrivateKey)
	if err != nil {
		n.logger.Errorf("Unable to sign tx: %s", err)
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	if err := n.AddPendingTX(signedTx, n.metadata); err != nil {
		n.logger.Errorf("Unable to add pending tx: %s", err)
		n.httpResponse(w, err, http.StatusInternalServerError)
		return
	}

	n.httpResponse(w, map[string]interface{}{
		"tx_hash": tx.Hash,
	})
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
	if bytes.Compare(hash.Bytes(), common.Hash{}.Bytes()) == 0 {
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
