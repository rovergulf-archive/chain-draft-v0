package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gorilla/mux"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/pkg/config"
	"github.com/rovergulf/rbn/wallets"
	"github.com/spf13/viper"
	"github.com/uber/jaeger-client-go"
	jconf "github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-lib/metrics/prometheus"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net"
)

const (
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

	config      config.Options
	bc          *core.Blockchain
	wm          *wallets.Manager
	httpHandler httpServer
	rpcListener net.Listener

	isMining bool

	knownPeers knownPeers

	proposedBlocks chan *core.Block
	proposedTXs    chan *core.Transaction
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
		config:     opts,
		bc:         nil,
		logger:     opts.Logger,
		knownPeers: make(map[string]PeerNode),
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
	nodeAddress := fmt.Sprintf("%s:%d", n.metadata.Ip, n.metadata.Port)

	httpApiAddress := fmt.Sprintf("%s:%s",
		viper.GetString("http.addr"),
		viper.GetString("http.port"),
	)

	n.logger.Infow("Starting node...",
		"addr", nodeAddress, "is_root", n.metadata.Root)

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

	ln, err := net.Listen(rpcNetProtocol, nodeAddress)
	if err != nil {
		return err
	} else {
		n.rpcListener = ln
	}

	if !n.metadata.Root {
	} else {
	}

	go func() {
		n.logger.Debugw("Listening TCP", "addr", nodeAddress)
		for {
			conn, err := n.rpcListener.Accept()
			if err != nil {
				n.logger.Errorf("Unable to accept connection from %s", conn.LocalAddr())
			}
			go func() {
				if err := n.HandleConnection(conn); err != nil {
					n.logger.Errorf("Unable to handle connection from %s", conn.LocalAddr())
				}
			}()
		}
	}()

	n.logger.Infow("Listening HTTP", "addr", httpApiAddress)
	return n.serveHttp()
}

func (n *Node) Shutdown() {
	n.bc.Shutdown()

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
	case "":
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
