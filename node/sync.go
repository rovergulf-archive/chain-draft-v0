package node

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

func (n *Node) sync(ctx context.Context) {
	n.doSync(ctx)

	syncTimerDuration := viper.GetDuration("node.sync_interval")
	ticker := time.NewTicker(syncTimerDuration * time.Second)

	for {
		select {
		case <-ticker.C:
			n.doSync(ctx)
		case <-ctx.Done():
			ticker.Stop()
		}
	}
}

func (n *Node) doSync(ctx context.Context) {
	for _, peer := range n.knownPeers.peers {
		if n.metadata.Ip == peer.Ip && n.metadata.Port == peer.Port {
			continue
		}

		if peer.Ip == "" {
			continue
		}

		n.logger.Infof("Searching for new Peers and their Blocks and Peers: '%s'", peer.TcpAddress())

		if err := n.joinKnownPeer(peer); err != nil {
			n.logger.Error("Unable to join known peer: ", err)
			continue
		}

		if err := n.syncBlocks(); err != nil {
			n.logger.Error("Unable to sync blocks: ", err)
			continue
		}

		if err := n.syncKnownPeers(); err != nil {
			n.logger.Error("Unable to sync knonw peers", err)
			continue
		}

		if err := n.syncPendingTXs(); err != nil {
			n.logger.Error(err)
			continue
		}
	}
}

func (n *Node) syncBlocks() error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncKnownPeers() error {
	for _, statusPeer := range n.knownPeers.peers {
		if !n.IsKnownPeer(statusPeer) {
			n.logger.Infof("Found new Peer %s", statusPeer.TcpAddress())

			if err := n.addPeer(statusPeer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) syncPendingTXs() error {
	return fmt.Errorf("not implemented")
}

func (n *Node) joinKnownPeer(peer PeerNode) error {

	return fmt.Errorf("not implemented")
}
