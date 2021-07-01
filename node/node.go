package node

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/database/badgerdb"
	"github.com/rovergulf/rbn/params"
	"github.com/rovergulf/rbn/pkg/exit"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	jconf "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.etcd.io/etcd/raft/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

const (
	DbFileName = "node.db"

	DefaultNodeIP      = "127.0.0.1"
	DefaultNodePort    = 9420
	HttpSSLPort        = 443
	DefaultNetworkAddr = "127.0.0.1:9420"
	rpcNetProtocol     = "tcp"

	endpointStatus  = "/node/status"
	endpointSync    = "/node/sync"
	endpointAddPeer = "/node/peer"

	CoinbaseAccount = "0x5793f98ea0911e12742c785c316903b50b0ddaca"
)

// Node represents blockchain network peer node
type Node struct {
	metadata PeerNode

	config params.Options

	bc *core.Blockchain
	wm *wallets.Manager
	db *badger.DB

	grpcServer  *grpc.Server
	httpHandler httpServer

	isMining bool

	knownPeers knownPeers

	pendingTXs map[string]core.SignedTx

	newSyncBlocks chan core.Block
	newSyncTXs    chan core.SignedTx

	raftStorage *raft.MemoryStorage
	//Lock *sync.RWMutex

	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
}

// New creates and returns new node if blockchain available
func New(opts params.Options) (*Node, error) {
	nodeAddr := viper.GetString("node.addr")
	nodePort := viper.GetUint64("node.port")

	syncMode := viper.GetString("node.sync_mode")
	if syncMode == "" {
		syncMode = string(SyncModeDefault)
	}

	mainNode := NewPeerNode(DefaultNodeIP, DefaultNodePort, common.HexToAddress(CoinbaseAccount), SyncMode(syncMode))
	pn := NewPeerNode(nodeAddr, nodePort, common.HexToAddress(opts.Address), SyncMode(syncMode))

	n := &Node{
		metadata: pn,
		httpHandler: httpServer{
			router: mux.NewRouter(),
			logger: opts.Logger,
		},
		config: opts,
		bc:     nil,
		logger: opts.Logger,
		knownPeers: map[string]PeerNode{
			mainNode.TcpAddress(): mainNode,
			pn.TcpAddress():       pn,
		},
		pendingTXs:    make(map[string]core.SignedTx),
		newSyncTXs:    make(chan core.SignedTx),
		newSyncBlocks: make(chan core.Block),
		//Lock:       new(sync.RWMutex),
	}

	jaegerTraceAddr := viper.GetString("jaeger_trace")
	if len(jaegerTraceAddr) > 0 {
		tracer, closer, err := n.initOpentracing(jaegerTraceAddr)
		if err != nil {
			return nil, err
		} else {
			n.tracer = tracer
			n.closer = closer
			n.httpHandler.tracer = tracer
		}
	}

	return n, nil
}

func (n *Node) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	nodeAddress := fmt.Sprintf("%s:%d", n.metadata.Ip, n.metadata.Port)

	n.logger.Infow("Starting node...",
		"addr", nodeAddress, "is_root", n.metadata.Root)

	exit.ListenExit(func(signal os.Signal) {
		n.logger.Warnf("Signal [%s] received. Graceful shutdown", signal)
		time.AfterFunc(15*time.Second, func() {
			n.logger.Fatal("Failed to gracefully shutdown after 15 sec. Force exit")
		})
		n.Shutdown()
	})

	db, err := badgerdb.OpenDB(viper.GetString("data_dir"), badger.DefaultOptions(n.config.NodeFilePath))
	if err != nil {
		n.logger.Errorf("Unable to open db file: %s", err)
		return err
	}
	n.db = db

	chain, err := core.ContinueBlockchain(n.config)
	if err != nil {
		return err
	} else {
		n.bc = chain
	}

	n.wm, err = wallets.NewManager(n.config)
	if err != nil {
		return err
	}

	n.logger.Debugf("Miner: %s", n.metadata.Account.Hex())

	if !n.metadata.Root {
	} else {
	}

	go func() {
		n.logger.Debugw("Listening gRPC", "addr", nodeAddress)
		grpcSrv, err := n.PrepareGrpcServer()
		if err != nil {
			n.logger.Errorf("Unable to prepare gRPC server: %s", err)
			n.Shutdown()
		}
		n.grpcServer = grpcSrv

		if err := n.RunGrpcServer(nodeAddress); err != nil {
			n.logger.Errorf("Unable to start gRPC server: %s", err)
			n.Shutdown()
		}
	}()

	go n.race(ctx)
	go n.sync(ctx)

	httpApiAddress := fmt.Sprintf("%s:%s",
		viper.GetString("http.addr"),
		viper.GetString("http.port"),
	)
	n.logger.Infow("Listening HTTP", "addr", httpApiAddress)
	return n.serveHttp()
}

func (n *Node) Shutdown() {

	var wg sync.WaitGroup

	if n.db != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.db.Close(); err != nil {
				n.logger.Errorf("Unable to close node db: %s", err)
			}
		}()
	}

	if n.bc != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.bc.Shutdown()
		}()
	}

	if n.wm != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.wm.Shutdown()
		}()
	}

	if n.grpcServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			n.grpcServer.GracefulStop()
		}()
	}

	if n.closer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := n.closer.Close(); err != nil {
				n.logger.Errorf("Unable to close tracing writer: %s", err)
			}
		}()
	}

	wg.Wait()

	os.Exit(0)
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	_, ok := n.knownPeers[peer.Account.Hex()]
	return ok
}

func (n *Node) initOpentracing(address string) (opentracing.Tracer, io.Closer, error) {
	metrics := prometheus.New()

	traceTransport, err := jaeger.NewUDPTransport(address, 0)
	if err != nil {
		n.logger.Errorf("Unable to setup tracing agent connection: %s", err)
		return nil, nil, err
	}

	tracer, closer, err := jconf.Configuration{
		ServiceName: "rbn",
	}.NewTracer(
		jconf.Sampler(jaeger.NewConstSampler(true)),
		jconf.Reporter(jaeger.NewRemoteReporter(
			traceTransport,
			jaeger.ReporterOptions.Logger(jaeger.StdLogger)),
		),
		jconf.Metrics(metrics),
	)
	if err != nil {
		n.logger.Errorf("Unable to start tracer: %s", err)
		return nil, nil, err
	}

	n.logger.Debugw("Jaeger tracing client initialized", "collector_url", address)
	return tracer, closer, nil
}

func (n *Node) collectNetInterfaces() error {
	ifaces, err := net.Interfaces()
	if err != nil {
		n.logger.Errorf("Unable to get net interfaces: %s", err)
		return err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			n.logger.Errorf("Unable to get net interface addrs: %s", err)
			return err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
				n.logger.Debugf("Discovered local IP network: %s", ip)
			case *net.IPAddr:
				ip = v.IP
				n.logger.Debugf("Discovered local IP address: %s", ip)
			}
		}
	}

	return nil
}

func (n *Node) race(ctx context.Context) {
	var miningCtx context.Context
	var stopCurrentMining context.CancelFunc

	ticker := time.NewTicker(time.Second * 10)

	for {
		select {
		case <-ticker.C:
			go func() {
				n.logger.Debugw("Check for available transactions",
					"txs", len(n.pendingTXs), "is_mining", n.isMining)
				if len(n.pendingTXs) > 0 && !n.isMining {
					n.isMining = true

					miningCtx, stopCurrentMining = context.WithCancel(ctx)
					// TODO rename
					if err := n.minePendingTXs(miningCtx); err != nil {
						n.logger.Errorf("Failed to mine pending txs: %s", err)
					}

					n.isMining = false
				}
			}()

		case block, _ := <-n.newSyncBlocks:
			n.logger.Debugw("Proposed block appeared", "is_mining", n.isMining)
			if n.isMining {
				n.logger.Warnf("Peer mined next Block '%s' faster :(", block.Hash.Hex())
				n.removeMinedPendingTXs(&block)
				stopCurrentMining()
			}

		case <-ctx.Done():
			ticker.Stop()
			n.logger.Debug("Mining context cancelled")
			break
		}
	}
}

func (n *Node) minePendingTXs(ctx context.Context) error {
	if len(n.pendingTXs) == 0 {
		return fmt.Errorf("no transactions available")
	}

	var txs []core.SignedTx

	for i := range n.pendingTXs {
		tx := n.pendingTXs[i]
		txs = append(txs, tx)
	}

	lb, err := n.bc.GetBlock(n.bc.LastHash)
	if err != nil {
		return err
	}

	header := core.BlockHeader{
		PrevHash:  lb.Hash,
		Number:    lb.Number + 1,
		Timestamp: time.Now().Unix(),
		Validator: common.Address{},
	}

	newBlock := core.NewBlock(header, txs)
	if err := newBlock.SetHash(); err != nil {
		return err
	}

	n.removeMinedPendingTXs(newBlock)

	if err := n.bc.AddBlock(newBlock); err != nil {
		return err
	}

	return nil
}

func (n *Node) removeMinedPendingTXs(block *core.Block) {
	if len(block.Transactions) > 0 && len(n.pendingTXs) > 0 {
		fmt.Println("Updating in-memory Pending TXs Pool:")
	}

	for _, tx := range block.Transactions {
		if _, exists := n.pendingTXs[tx.Hash.Hex()]; exists {
			fmt.Printf("\t-archiving mined TX: %s\n", tx.Hash.Hex())

			delete(n.pendingTXs, tx.Hash.Hex())
		}
	}
}
func (n *Node) AddPendingTX(tx core.SignedTx, peer PeerNode) error {
	//ok, err := tx.IsAuthentic()
	//if err != nil {
	//	return err
	//}
	//
	//if !ok {
	//	return fmt.Errorf("wrong TX. Sender '%s' is forged", tx.From)
	//}

	if err := n.bc.SaveTx(tx); err != nil {
		return err
	}

	if err := n.bc.ApplyTx(tx); err != nil {
		return err
	}

	n.pendingTXs[tx.Hash.Hex()] = tx

	return nil
}
