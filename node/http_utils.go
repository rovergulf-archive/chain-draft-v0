package node

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/resutil"
	"go.uber.org/zap"
	"net/http"
	"strings"
)

const (
	endpointSyncQueryKeyFromBlock = "fromBlock"
	endpointAddPeerQueryKeyIP     = "ip"
	endpointAddPeerQueryKeyPort   = "port"
	endpointAddPeerQueryKeyMiner  = "miner"
)

var allowedHeaders = []string{
	"Accept",
	"Content-Type",
	"Content-Length",
	"Cookie",
	"Accept-Encoding",
	"Authorization",
	"X-CSRF-Token",
	"X-Requested-With",
	"X-Node-ID",
}

var allowedMethods = []string{
	"OPTIONS",
	"GET",
	"PUT",
	"PATCH",
	"POST",
	"DELETE",
}

// httpServer represents mux.Router interceptor, to handle CORS requests
type httpServer struct {
	router *mux.Router
	tracer opentracing.Tracer
	logger *zap.SugaredLogger
}

// ServeHTTP wraps http.Server ServeHTTP method to handle preflight requests
func (h *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Set request headers for AJAX requests
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
	}

	// handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if h.tracer != nil {
		span := h.tracer.StartSpan(strings.TrimPrefix(r.URL.Path, "/"))
		span.SetTag("host", r.Host)
		span.SetTag("method", r.Method)
		span.SetTag("path", r.URL.Path)
		span.SetTag("query", r.URL.RawQuery)
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	h.logger.Infow("Handling request", "method", r.Method, "path", r.URL.Path, "query", r.URL.RawQuery)

	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func (n *Node) httpResponse(w http.ResponseWriter, i interface{}, statusCode ...int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if len(statusCode) > 0 {
		w.WriteHeader(statusCode[0])
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := resutil.WriteJSON(w, n.logger, i); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Unable to write json response: %s", err)))
	}
}

type StatusRes struct {
	LastHash   string                   `json:"block_hash,omitempty" yaml:"last_hash,omitempty"`
	Number     uint64                   `json:"chain_length,omitempty" yaml:"chain_length,omitempty"`
	KnownPeers map[string]PeerNode      `json:"peers_known,omitempty" yaml:"known_peers,omitempty"`
	PendingTXs map[string]core.SignedTx `json:"pending_txs,omitempty" yaml:"pending_t_xs,omitempty"`
	IsMining   bool                     `json:"is_mining" yaml:"is_mining"`
	DbSize     map[string]int64         `json:"db_size" yaml:"db_size"`
}

type SyncRes struct {
	Blocks []*core.Block `json:"blocks" yaml:"blocks"`
}

type AddPeerRes struct {
	Success bool   `json:"success" yaml:"success"`
	Error   string `json:"error,omitempty" yaml:"error,omitempty"`
}
