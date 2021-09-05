package node

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	tls "github.com/libp2p/go-libp2p-tls"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-libp2p/p2p/discovery"
	"github.com/libp2p/go-tcp-transport"
	websocket "github.com/libp2p/go-ws-transport"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
	"time"
)

type mdnsNotifee struct {
	logger *zap.SugaredLogger
	h      host.Host
	ctx    context.Context
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if err := m.h.Connect(m.ctx, pi); err != nil {
		m.logger.Errorw("Unable to join peer", "peer_id", pi.ID, "err", err)
	}
}

func (n *Node) PrepareP2pPeer(ctx context.Context) (host.Host, error) {
	transports := libp2p.ChainOptions(
		libp2p.Transport(tcp.NewTCPTransport), // this refers to multiaddrs usage
		libp2p.Transport(websocket.New),       // same, as one above
	)

	muxers := libp2p.ChainOptions(
		libp2p.Muxer("/yamux/1.0.0", yamux.DefaultTransport),
		libp2p.Muxer("/mplex/6.7.0", mplex.DefaultTransport),
	)

	security := libp2p.Security(tls.ID, tls.New)

	listenAddrs := libp2p.ListenAddrStrings(
		"/ip4/0.0.0.0/tcp/0",
		"/ip4/0.0.0.0/tcp/0/ws",
	)

	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		if n.dht, err = kaddht.New(ctx, h); err != nil {
			n.logger.Errorf("Unable to create dht: %s", err)
			return nil, err
		} else {
			n.logger.Debugw("New DHT ipfs peer", "peer_id", n.dht.PeerID())
		}
		return n.dht, nil
	}
	router := libp2p.Routing(newDHT)

	h, err := libp2p.New(
		ctx,
		transports,
		listenAddrs,
		muxers,
		security,
		router,
	)
	if err != nil {
		n.logger.Errorf("Unable to create p2p host: %s", err)
		return nil, err
	}

	return h, nil
}

// PrepareSubs initializes and runs main topic subscription
func (n *Node) PrepareSubs(ctx context.Context, channel string) error {
	ps, err := pubsub.NewGossipSub(ctx, n.host)
	if err != nil {
		n.logger.Errorf("Unable to create gossip sub: %s", err)
		return err
	}

	n.mainTopic, err = ps.Join(channel)
	if err != nil {
		n.logger.Errorf("Unable to join topic '%s': %s", channel, err)
		return err
	}

	n.mainSub, err = n.mainTopic.Subscribe()
	if err != nil {
		n.logger.Errorf("Unable to subscribe to '%s': %s", channel, err)
		return err
	}

	return nil
}

// RunP2pServer runs peer node
func (n *Node) RunP2pServer(ctx context.Context) error {

	for _, addr := range n.host.Addrs() {
		n.logger.Debugf("Listening on '%s'", addr)
	}

	targetAddr, err := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/63785/p2p/QmWjz6xb8v9K4KnYEwP5Yk75k5mMBCehzWFLCvvQpYxF3d")
	if err != nil {
		n.logger.Errorf("Failed to connect target: %s", err)
		return err
	}

	targetInfo, err := peer.AddrInfoFromP2pAddr(targetAddr)
	if err != nil {
		n.logger.Errorf("Unable to get target info: %s", err)
		return err
	}

	n.logger.Infof("Connected to '%s'", targetInfo.ID)

	mdns, err := discovery.NewMdnsService(ctx, n.host, 5*time.Second, "")
	if err != nil {
		n.logger.Errorf("Unable to create new mdns service: %s", err)
		return err
	}
	mdns.RegisterNotifee(&mdnsNotifee{h: n.host, ctx: ctx})

	if err := n.dht.Bootstrap(ctx); err != nil {
		n.logger.Errorf("Failed to bootstrap dht: %s", err)
		return err
	}

	return nil
}
