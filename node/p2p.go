package node

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/rovergulf/rbn/params"
	"github.com/spf13/viper"
)

func (n *Node) newEthP2pServer(ctx context.Context) error {
	listenAddr := fmt.Sprintf("%s:%s", viper.GetString("node.addr"), viper.GetString("node.port"))
	config := p2p.Config{
		Name:             common.MakeName("Nether Node", params.Version),
		MaxPeers:         256,
		ListenAddr:       listenAddr,
		DiscoveryV5:      true,
		PrivateKey:       n.account.GetKey().PrivateKey,
		TrustedNodes:     n.getTrustedNodes(),
		BootstrapNodesV5: n.getBootstrapNodes(),
		Protocols:        []p2p.Protocol{},
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
	nodes = append(nodes, enode.MustParseV4(params.LocalNodeAddr))
	// temp second node
	nodes = append(nodes, enode.MustParseV4("enode://8a83023555d2cbadf5c8f34b77fe6687fce576b7747241f17eced939ab713a00039ed605dc53bce1b8ece741c5cc509741d7963eee097d0fb06847f978577c09@127.0.0.1:9421"))
	return nodes
}

func (n *Node) getServerProtocols() []p2p.Protocol {
	var proto []p2p.Protocol
	return proto
}

func (n *Node) connectPeers(ctx context.Context) error {

	return nil
}
