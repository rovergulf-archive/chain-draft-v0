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
		Length:  1,
		Run: func(peer *p2p.Peer, rw p2p.MsgReadWriter) error {
			n.logger.Infow("Run peer proto", "id", peer.ID(), "local_addr", peer.LocalAddr())
			return nil
		},
		//NodeInfo: func() interface{} {
		//	n.logger.Info("protocol node info")
		//	return n.srv.NodeInfo()
		//},
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
