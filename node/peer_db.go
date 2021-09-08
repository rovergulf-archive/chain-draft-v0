package node

import (
	"context"
	"github.com/dgraph-io/badger/v3"
)

// addPeer saves new peer to node storage
func (n *Node) addDbPeer(ctx context.Context, peer PeerNode) error {
	pn, err := peer.Serialize()
	if err != nil {
		return err
	}

	key := append(peerPrefix, []byte(peer.id)...)
	return n.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, pn)
	})
}

// removePeer deletes peer from node storage
func (n *Node) removeDbPeer(ctx context.Context, peer PeerNode) error {
	return n.db.Update(func(txn *badger.Txn) error {
		key := append(peerPrefix, []byte(peer.id)...)
		return txn.Delete(key)
	})
}

func (n *Node) searchPeers(ctx context.Context) ([]*PeerNode, error) {
	var peers []*PeerNode

	if err := n.db.View(func(txn *badger.Txn) error {
		return nil
	}); err != nil {
		return nil, err
	}

	return peers, nil
}
