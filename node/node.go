package node

import (
	"context"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/rovergulf/rbn/pkg/exit"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	jconf "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
	"os"
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

	RootAddress  = "0x59fc6df01d2e84657faba24dc96e14871192bda4"
	DefaultMiner = "0x0000000000000000000000000000000000000000"
)

// Node represents blockchain network peer node
type Node struct {
	metadata PeerNode

	config config.Options

	bc *core.Blockchain
	wm *wallets.Manager
	db *badger.DB

	httpHandler httpServer
	rpcListener net.Listener

	isMining bool

	knownPeers knownPeers

	pendingTXs map[string]*core.SignedTx

	proposedBlocks chan *core.Block
	proposedTXs    chan *core.SignedTx
	errCh          chan error

	//Lock *sync.RWMutex

	logger *zap.SugaredLogger
	tracer opentracing.Tracer
	closer io.Closer
}

// New creates and returns new node if blockchain available
func New(opts config.Options) (*Node, error) {
	nodeAddr := viper.GetString("node.addr")
	nodePort := viper.GetUint64("node.port")
	peerNodeAddr := fmt.Sprintf("%s:%d", nodeAddr, nodePort)

	pn := PeerNode{
		Ip:        nodeAddr,
		Port:      nodePort,
		Root:      peerNodeAddr == DefaultNetworkAddr,
		Account:   common.HexToAddress(opts.Address),
		connected: false,
	}

	n := &Node{
		metadata: pn,
		httpHandler: httpServer{
			router: mux.NewRouter(),
			logger: opts.Logger,
		},
		config:         opts,
		bc:             nil,
		logger:         opts.Logger,
		knownPeers:     make(map[string]PeerNode),
		pendingTXs:     make(map[string]*core.SignedTx),
		proposedTXs:    make(chan *core.SignedTx),
		proposedBlocks: make(chan *core.Block),
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

	httpApiAddress := fmt.Sprintf("%s:%s",
		viper.GetString("http.addr"),
		viper.GetString("http.port"),
	)

	n.logger.Infow("Starting node...",
		"addr", nodeAddress, "is_root", n.metadata.Root)

	exit.ListenExit(func(signal os.Signal) {
		n.logger.Warnf("Signal [%s] received. Graceful shutdown", signal)
		time.AfterFunc(15*time.Second, func() {
			n.logger.Fatal("Failed to gracefully shutdown after 15 sec. Force exit")
		})
		n.Shutdown()
	})

	chain, err := core.ContinueBlockchain(n.config)
	if err != nil {
		return err
	}
	defer chain.Shutdown()
	n.bc = chain

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

	}()

	go n.mine(ctx)

	n.logger.Infow("Listening HTTP", "addr", httpApiAddress)
	return n.serveHttp()
}

func (n *Node) Shutdown() {
	if n.bc != nil {
		n.bc.Shutdown()
	}

	if n.wm != nil {
		n.wm.Shutdown()
	}

	if n.rpcListener != nil {
		if err := n.rpcListener.Close(); err != nil {
			n.logger.Errorf("Unable to close rpc listener: %s", err)
		}
	}

	if n.closer != nil {
		if err := n.closer.Close(); err != nil {
			n.logger.Errorf("Unable to close tracing writer: %s", err)
		}
	}

	os.Exit(0)
}

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	_, ok := n.knownPeers[peer.Account.Hex()]
	return ok
}

func (n *Node) HandleConnection(conn net.Conn) error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	defer conn.Close()

	req, err := ioutil.ReadAll(conn)
	if err != nil {
		return err
	}

	command := BytesToCmd(req[:12])

	if n.tracer != nil {
		span := n.tracer.StartSpan("node_rpc_conn")
		span.SetTag("cmd", command)
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	n.logger.Debugf("Received [%s] command", command)

	switch command {
	case "sync":
		return nil
		//return HandleAddr(req)
	case "block":
		return nil
		//return HandleBlock(req, chain)
	case "inv":
		return nil
		//return HandleInv(req, chain)
	case "getblocks":
		return nil
		//return HandleGetBlocks(req, chain)
	case "getdata":
		return nil
		//return HandleGetData(req, chain)
	case "tx":
		return nil
		//return HandleTx(req, chain)
	case "version":
		return nil
		//return HandleVersion(req, chain)
	default:
		return fmt.Errorf("unknown command")
	}
}

func CmdToBytes(cmd string) []byte {
	var bytes [12]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
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

func (n *Node) mine(ctx context.Context) error {
	var miningCtx context.Context
	var stopCurrentMining context.CancelFunc

	ticker := time.NewTicker(time.Second * miningIntervalSeconds)

	for {
		select {
		case <-ticker.C:
			go func() {
				n.logger.Debugw("Check for available transactions",
					"txs", len(n.proposedTXs), "is_mining", n.isMining)
				if len(n.proposedTXs) > 0 && !n.isMining {
					n.isMining = true

					miningCtx, stopCurrentMining = context.WithCancel(ctx)
					err := n.minePendingTXs(miningCtx)
					if err != nil {
						n.logger.Errorf("Failed to mine pending txs: %s", err)
					}

					n.isMining = false
				}
			}()

		case block, _ := <-n.proposedBlocks:
			n.logger.Debugw("Proposed block appeared", "is_mining", n.isMining)
			if n.isMining {
				n.logger.Warnf("Peer mined next Block '%s' faster :(", block.Hash.Hex())
				n.removeMinedPendingTXs(block)
				stopCurrentMining()
			}

		case <-ctx.Done():
			ticker.Stop()
			n.logger.Debug("Mining context cancelled")
			return nil
		}
	}
}

func (n *Node) minePendingTXs(ctx context.Context) error {
	if len(n.pendingTXs) == 0 {
		return fmt.Errorf("no transactions available")
	}

	var txs []*core.SignedTx

	for i := range n.pendingTXs {
		txs = append(txs, n.pendingTXs[i])
	}

	blockToMine := NewPendingBlock(
		n.bc.LastHash,
		n.bc.ChainLength.Uint64(),
		n.metadata.Account,
		txs,
	)

	minedBlock, err := Mine(ctx, blockToMine)
	if err != nil {
		return err
	}

	n.removeMinedPendingTXs(minedBlock)

	if err := n.bc.AddBlock(minedBlock); err != nil {
		return err
	}

	return nil
}

func (n *Node) removeMinedPendingTXs(block *core.Block) {
	if len(block.Transactions) > 0 && len(n.pendingTXs) > 0 {
		fmt.Println("Updating in-memory Pending TXs Pool:")
	}

	for _, tx := range block.Transactions {
		txHash, _ := tx.Hash()
		txh := common.BytesToHash(txHash)
		if _, exists := n.pendingTXs[txh.Hex()]; exists {
			fmt.Printf("\t-archiving mined TX: %s\n", txh.Hex())

			delete(n.pendingTXs, txh.Hex())
		}
	}
}
