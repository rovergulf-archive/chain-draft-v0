package rpc

import (
	"context"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/node"
	"github.com/rovergulf/rbn/rpc/pb"
	"github.com/rovergulf/rbn/wallets"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Namespace string      `json:"namespace" yaml:"namespace"` // namespace under which the rpc methods of Service are exposed
	Version   string      `json:"version" yaml:"version"`     // api version for DApp's
	Service   interface{} `json:"service" yaml:"service"`     // receiver instance which holds the methods
	Public    bool        `json:"public" yaml:"public"`       // indication if the methods must be considered safe for public use
}

type Stack struct {
	Apis []API `json:"apis" yaml:"apis"`

	handlers map[pb.Command]map[pb.Entity]ApiHandler

	bc  *core.BlockChain
	ndb *node.DB
	wm  *wallets.Manager
}

type ApiHandler func(ctx context.Context, req *pb.CallRequest) (*pb.CallResponse, error)
