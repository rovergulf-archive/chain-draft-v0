package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/params"
	"github.com/spf13/viper"
	"io/ioutil"
	"sync/atomic"
	"time"
)

func (n *Node) newEthP2pServer(ctx context.Context) error {
	listenAddr := fmt.Sprintf("%s:%s", viper.GetString("node.addr"), viper.GetString("node.port"))
	config := p2p.Config{
		Name:           common.MakeName("Nether Node", params.Version),
		MaxPeers:       256,
		ListenAddr:     listenAddr,
		DiscoveryV5:    true,
		PrivateKey:     n.account.GetKey().PrivateKey,
		BootstrapNodes: mainNetBootNodes(),
		Protocols:      n.getServerProtocols(),
	}

	n.srv = &p2p.Server{
		Config: config,
	}

	return n.srv.Start()
}

func (n *Node) peerFunc(peer *p2p.Peer) {
	n.logger.Infow("peerFunc", "id", peer.ID(), "info", peer.Info())
}

func (n *Node) newEthP2pPeer(ctx context.Context) error {
	return nil
}

func (n *Node) getBootstrapNodes() []*enode.Node {
	var nodes []*enode.Node
	return nodes
}

func (n *Node) getTrustedNodes() []*enode.Node {
	var nodes []*enode.Node
	// temp second node

	return nodes
}

func (n *Node) getServerProtocols() []p2p.Protocol {
	var protos []p2p.Protocol
	protos = append(protos, p2p.Protocol{
		Name:    "rbn",
		Version: 1,
		Length:  2,
		Run: func(p *p2p.Peer, rw p2p.MsgReadWriter) error {
			peer := NewPeer(p, rw)
			defer peer.Close()

			go n.announceTx(peer)
			go n.announceBlocks(peer)

			n.logger.Infow("New peer", "id", peer.id)
			return n.runPeer(peer)
		},
		NodeInfo: n.Info,
		//PeerInfo: func(id enode.ID) interface{} {
		//	n.logger.Infof("protocol enode id: %s", id.String())
		//	return n.metadata
		//},
		DialCandidates: nil,
		Attributes:     nil,
	})
	return protos
}

func (n *Node) connectPeers(ctx context.Context) error {

	return nil
}

const (
	StatusMsg = iota
	NewBlockHashesMsg
	TransactionsMsg
	GetBlockHeadersMsg
	BlockHeadersMsg
	GetBlockBodiesMsg
	BlockBodiesMsg
	NewBlockMsg
	GetNodeDataMsg
	NodeDataMsg
	GetReceiptsMsg
	ReceiptsMsg
	NewPooledTransactionHashesMsg
	GetPooledTransactionsMsg
	PooledTransactionsMsg
)

var (
	handshakeTimeout = 5 * time.Second
)

func (n *Node) runPeer(p *Peer) error {
	ctx := context.Background()

	var span opentracing.Span
	if n.tracer != nil {
		span = n.tracer.StartSpan("handle_peer")
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	msg, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}

	if span != nil {
		span.SetTag("msg_code", msg.Code)
		span.SetBaggageItem("ack", "true")
	}

	payload, err := ioutil.ReadAll(msg.Payload)
	if err != nil {
		return err
	}

	if span != nil {
		span.SetBaggageItem("read", "true")
	}

	var res *CallResult
	switch msg.Code {
	case NewBlockHashesMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case TransactionsMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case GetBlockHeadersMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case BlockHeadersMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case GetBlockBodiesMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case BlockBodiesMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case NewBlockMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case GetNodeDataMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case NodeDataMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case GetReceiptsMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case ReceiptsMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case NewPooledTransactionHashesMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case GetPooledTransactionsMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	case PooledTransactionsMsg:
		if res, err = nilPeerHandler(ctx, payload); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid message code")
	}

	if res != nil {
		if err := p2p.Send(p.rw, res.Code, res.Data); err != nil {
			n.logger.Errorw("Unable to send p2p message", "err", err)
			return err
		}
	}

	return nil
}

func (n *Node) handshake(ctx context.Context, p *Peer) error {
	errC := make(chan error, 2)

	var peerStatus StatusResult

	go func() {
		errC <- p2p.Send(p.rw, StatusMsg, StatusResult{
			Head:      "",
			Genesis:   "",
			NetworkId: "",
			Uptime:    0,
		})
	}()

	go func() {
		errC <- n.readStatus(ctx, p, &peerStatus)
	}()

	timeout := time.NewTimer(handshakeTimeout)
	defer timeout.Stop()
	for i := 0; i < 2; i++ {
		select {
		case err := <-errC:
			if err != nil {
				return err
			}
		case <-timeout.C:
			return p2p.DiscReadTimeout
		}
	}

	return nil
}

func (n *Node) readStatus(ctx context.Context, p *Peer, status *StatusResult) error {

	return nil
}

func (n *Node) announceBlocks(p *Peer) {
	for {
		select {
		case nb := <-n.blockBroadcast:
			n.logger.Infow("new broadcast block", "hash", nb.BlockHash)
		case ab := <-n.blockAnnounce:
			n.logger.Infow("new announce block", "hash", ab.BlockHash)
		}
	}
}

func (n *Node) announceTx(p *Peer) {
	for {
		select {
		case btx := <-n.txBroadcast:
			n.logger.Infow("new broadcast block", "txs count", len(btx))
		case atx := <-n.txAnnounce:
			n.logger.Infow("new announce block", "txs count", len(atx))
		}
	}
}

func (n *Node) Info() interface{} {
	return struct {
		Received int64 `json:"received"`
	}{
		atomic.LoadInt64(&n.received),
	}
}

func (n *Node) PeerInfo(id enode.ID) interface{} {
	return nil
}

func nilPeerHandler(ctx context.Context, payload []byte) (*CallResult, error) {
	return nil, fmt.Errorf("not implemented")
}
