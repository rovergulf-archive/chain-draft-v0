package node

import (
	"context"
	"net/http"
)

func (n *Node) wsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		n.httpHandler.ServeHTTP(w, r.WithContext(ctx))
	})
}
