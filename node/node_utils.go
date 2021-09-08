package node

import (
	"context"
)

func (n *Node) IsKnownPeer(peer PeerNode) bool {
	_, ok := n.knownPeers.GetPeer(peer.TcpAddress())
	return ok
}

func (n *Node) removeKnownPeer(ctx context.Context, peer PeerNode) error {
	if err := n.removeDbPeer(ctx, peer); err != nil {
		return err
	}

	n.knownPeers.DeletePeer(peer.TcpAddress())
	return nil
}
