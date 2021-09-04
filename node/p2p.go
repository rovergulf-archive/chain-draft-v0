package node

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/routing"
	kaddht "github.com/libp2p/go-libp2p-kad-dht"
	mplex "github.com/libp2p/go-libp2p-mplex"
	tls "github.com/libp2p/go-libp2p-tls"
	yamux "github.com/libp2p/go-libp2p-yamux"
	"github.com/libp2p/go-tcp-transport"
	websocket "github.com/libp2p/go-ws-transport"
)

type mdnsNotifee struct {
	h   host.Host
	ctx context.Context
}

func (m *mdnsNotifee) HandlePeerFound(pi peer.AddrInfo) error {
	if err := m.h.Connect(m.ctx, pi); err != nil {
		return err
	}

	return nil
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

	var dht *kaddht.IpfsDHT
	newDHT := func(h host.Host) (routing.PeerRouting, error) {
		var err error
		if dht, err = kaddht.New(ctx, h); err != nil {
			n.logger.Errorf("Unable to create dht: %s", err)
			return nil, err
		} else {
			n.logger.Debugw("New DHT ipfs peer", "peer_id", dht.PeerID())
		}
		return dht, nil
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

func (n *Node) RunP2pServer(ctx context.Context, addr string) error {

	return nil
}
