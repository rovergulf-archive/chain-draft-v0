package node

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/gorilla/mux"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/core/types"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
	"github.com/rovergulf/rbn/pkg/sigutil"
	"github.com/rovergulf/rbn/pkg/traceutil"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"os"
	"time"
)

const (
	DbFileName = "node.db"

	DefaultNodeIP   = "127.0.0.1"
	DefaultNodePort = 9420
	HttpSSLPort     = 443

	endpointStatus  = "/node/status"
	endpointSync    = "/node/sync"
	endpointAddPeer = "/node/peer"
)

// Node represents blockchain network peer node
type Node struct {
	metadata PeerNode
	account  *wallets.Wallet

	config params.Options

	bc *core.BlockChain
	wm *wallets.Manager
	db *badger.DB

	httpHandler httpServer //

	srv *p2p.Server // eth p2p server instance

	inGenRace bool

	knownPeers knownPeers

	// network state
	pendingState *pendingState

	newSyncBlocks chan types.Block    // ??
	newSyncTXs    chan types.SignedTx // ??

	blockBroadcast chan types.Block       // ??
	blockAnnounce  chan types.BlockHeader // ??

	txBroadcast chan []common.Hash // ??
	txAnnounce  chan []common.Hash // ??

	//Lock *sync.RWMutex
	received int64

	// utils
	logger *zap.SugaredLogger
	tracer traceutil.Tracer
}

// New creates and returns new node if blockchain available
func New(opts params.Options) (*Node, error) {
	n := &Node{
		httpHandler: httpServer{
			router: mux.NewRouter(),
			logger: opts.Logger,
			tracer: opts.Tracer,
		},
		config:         opts,
		logger:         opts.Logger,
		blockBroadcast: make(chan types.Block),
		blockAnnounce:  make(chan types.BlockHeader),
		txBroadcast:    make(chan []common.Hash),
		txAnnounce:     make(chan []common.Hash),
	}

	return n, nil
}

func (n *Node) Init(ctx context.Context) error {
	sigutil.ListenExit(func(signal os.Signal) {
		n.logger.Warnf("Signal [%s] received. Graceful shutdown initialized.", signal)
		time.AfterFunc(15*time.Second, func() {
			n.logger.Fatal("Failed to gracefully shutdown after 15 sec. Force exit")
		})
		n.Shutdown()
	})

	tracer, err := traceutil.NewTracerFromViperConfig()
	if err != nil {
		if err != traceutil.ErrCollectorUrlNotSpecified {
			return err
		}
	} else {
		n.tracer = tracer
		n.httpHandler.tracer = tracer
	}

	db, err := badgerdb.OpenDB(viper.GetString("data_dir"), badger.DefaultOptions(n.config.NodeFilePath))
	if err != nil {
		n.logger.Errorf("Unable to open db file: %s", err)
		return err
	}
	n.db = db

	chain, err := core.NewBlockChain(n.config)
	if err != nil {
		n.logger.Errorf("Unable to continue blockchain: %s", err)
		return err
	} else {
		n.bc = chain
	}

	if err := chain.LoadChainState(ctx); err != nil {
		n.logger.Errorf("Unable to continue blockchain: %s", err)
		return err
	}

	n.wm, err = wallets.NewManager(n.config)
	if err != nil {
		n.logger.Errorf("Unable to init wallets manager: %s", err)
		return err
	}

	if err := n.setupNodeAccount(); err != nil {
		n.logger.Errorf("Unable to setup node account: %s", err)
		return err
	}
	n.logger.Debugf("Node account: %s", n.account.Address())

	//if err := n.syncKeystoreBalances(ctx); err != nil {
	//	n.logger.Errorf("Unable to sync keystore balances with chain state: %s", err)
	//}

	return nil
}

func (n *Node) Run(ctx context.Context) error {
	go func() {
		nodeAddress := fmt.Sprintf("%s:%d",
			viper.GetString("node.addr"), viper.GetInt("node.port"))
		n.logger.Debugw("Starting p2p node", "addr", nodeAddress)
		if err := n.newEthP2pServer(ctx); err != nil {
			n.logger.Errorf("Unable to run eth.p2p server: %s", err)
			n.Shutdown()
		}
	}()

	httpApiAddress := fmt.Sprintf("%s:%s",
		viper.GetString("http.addr"),
		viper.GetString("http.port"),
	)
	n.logger.Infow("Start listening HTTP", "addr", httpApiAddress)
	return n.serveHttp()
}

func (n *Node) Shutdown() {
	//close(n.newSyncTXs)
	//close(n.newSyncBlocks)

	if n.srv != nil {
		n.srv.Stop()
	}

	if n.db != nil {
		if err := n.db.Close(); err != nil {
			n.logger.Errorf("Unable to close node db: %s", err)
		}
	}

	if n.bc != nil {
		n.bc.Shutdown()
	}

	if n.wm != nil {
		n.wm.Shutdown()
	}

	if n.tracer != nil {
		n.tracer.Close()
	}

	os.Exit(0)
}
