package rpc

import (
	"context"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string `json:"namespace" yaml:"namespace"` // namespace under which the rpc methods of Service are exposed
	Version   string `json:"version" yaml:"version"`     // api version for DApp's
	Public    bool   `json:"public" yaml:"public"`       // indication if the methods must be considered safe for public use
}

type Stack struct {
	Apis []API `json:"apis" yaml:"apis"`

	handlers map[int]ApiHandler

	// cannot be used, circular dependency guaranteed
	// github issue: https://github.com/rovergulf/chain/issues/30
	//bc  *core.BlockChain
	//ndb *node.DB
	//wm  *wallets.Manager
}

type ApiRequest struct {
	Code      uint64 `json:"code" yaml:"code"`
	Namespace string `json:"namespace" yaml:"namespace"`
	PeerId    string `json:"peer_id" yaml:"peer_id"`
	Data      []byte `json:"data" yaml:"data"`
}

type ApiResponse struct {
	Code uint64 `json:"code" yaml:"code"`
	Data []byte `json:"data" yaml:"data"`
}

type ApiHandler func(ctx context.Context, req *ApiRequest) (*ApiResponse, error)
