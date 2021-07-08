package node

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/prom2json"
	"github.com/rovergulf/rbn/params"
	"github.com/rovergulf/rbn/pkg/resutil"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
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
	router     *mux.Router
	tracer     opentracing.Tracer
	logger     *zap.SugaredLogger
	httpServer http.Server
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
	w.Header().Set("Content-Type", "application/json")
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

func (n *Node) healthCheck(w http.ResponseWriter, r *http.Request) {
	n.httpResponse(w, map[string]interface{}{
		"http_status": http.StatusOK,
		"timestamp":   time.Now().Unix(),
		"run_date":    params.RunDate.Format(time.RFC1123),
		"node_status": "healthy",
		"in_gamble":   n.inGenRace,
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
