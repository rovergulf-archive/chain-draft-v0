package node

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	DefaultIP          = "127.0.0.1"
	HttpSSLPort        = 443
	DefaultNetworkAddr = "127.0.0.1:9420"

	endpointStatus  = "/node/status"
	endpointSync    = "/node/sync"
	endpointAddPeer = "/node/peer"
)

// Node represents blockchain network peer node
type Node struct {
	metadata PeerNode

	config      config.Options
	bc          *core.Blockchain
	logger      *zap.SugaredLogger
	httpHandler httpServer

	isMining   bool
	knownPeers map[string]PeerNode

	newSyncedBlocks chan *core.Block
	newPendingTXs   chan *core.Transaction

	//Lock *sync.RWMutex
}

// New creates and returns new node if blockchain available
func New(opts config.Options) (*Node, error) {
	nodeAddr := viper.GetString("node.addr")
	nodePort := viper.GetUint64("node.port")
	peerNodeAddr := fmt.Sprintf("%s:%d", nodeAddr, nodePort)

	n := &Node{
		metadata: PeerNode{
			Ip:        nodeAddr,
			Port:      nodePort,
			Root:      peerNodeAddr == DefaultNetworkAddr,
			Account:   opts.Address,
			connected: false,
		},
		httpHandler: httpServer{
			router: mux.NewRouter(),
			tracer: opts.Tracer,
			logger: opts.Logger,
		},
		config:     opts,
		bc:         nil,
		logger:     opts.Logger,
		knownPeers: make(map[string]PeerNode),
		//Lock:       new(sync.RWMutex),
	}

	return n, nil
}

func (n *Node) Run() error {
	nodeAddress := fmt.Sprintf("%s:%s",
		viper.GetString("node.addr"),
		viper.GetString("node.port"),
	)

	n.logger.Infow("Starting node server",
		"addr", nodeAddress, "is_root", n.metadata.Root)

	chain, err := core.ContinueBlockchain(n.config)
	if err != nil {
		return err
	}
	defer chain.Shutdown()
	n.bc = chain

	if !n.metadata.Root {
		//if err := SendVersion(KnownNodes[0], chain); err != nil {
		//	return err
		//}
	} else {

	}

	return n.serveHttp()
}

func (n *Node) Shutdown() {
	n.bc.Shutdown()

}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	_, ok := n.knownPeers[peer.GetId()]
	return ok
}
