package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rovergulf/rbn/client"
	"github.com/rovergulf/rbn/node/pb"
	"github.com/spf13/viper"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
			n.logger.Debug("Removing itself from known peers mem cache")
			n.knownPeers.DeletePeer(peer.TcpAddress())
			continue
		}

		if peer.Ip == "" {
			n.logger.Debug("Removing node with no IP Address from known peers mem cache")
			n.knownPeers.DeletePeer(peer.TcpAddress())
			continue
		}

		n.logger.Infof("Searching for new Peers and their Blocks and Peers: '%s'", peer.TcpAddress())

		n.logger.Debug("Join known peers...")
		if err := n.joinKnownPeer(ctx, peer); err != nil {
			n.logger.Error("Unable to join known peer: ", err)
			continue
		}

		//n.logger.Debug("Sync state version...")
		//if err := n.syncVersion(ctx); err != nil {
		//	n.logger.Error("Unable to sync version: ", err)
		//	continue
		//}

		//n.logger.Debug("Validate genesis...")
		//if err := n.validateGenesis(ctx); err != nil {
		//	n.logger.Error("Unable to validate genesis: ", err)
		//	continue
		//}

		n.logger.Debug("Sync state account balances...")
		if err := n.syncAccountBalances(ctx, n.metadata); err != nil {
			n.logger.Error("Unable to sync version: ", err)
			continue
		}

		//n.logger.Debug("Sync blocks...")
		//if err := n.syncBlocks(ctx); err != nil {
		//	n.logger.Error("Unable to sync blocks: ", err)
		//	continue
		//}

		//n.logger.Debug("Sync pending transactions...")
		//if err := n.syncPendingTXs(ctx); err != nil {
		//	n.logger.Error(err)
		//	continue
		//}
	}
}

func (n *Node) validateGenesis(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) joinKnownPeer(ctx context.Context, peer PeerNode) error {
	if n.IsKnownPeer(peer) {
		n.logger.Debugw("Peer already known",
			"account", peer.Account, "addr", peer.TcpAddress())
		return nil
	}

	if !peer.connected {
		// TODO client.NewClient
		c, err := client.NewClient(ctx, n.logger, peer.TcpAddress())
		if err != nil {
			n.logger.Errorf("Unable to run peer client: %s", err)
			return err
		}

		peer.connected = true
		peer.client = c
		n.knownPeers.AddPeer(peer.TcpAddress(), peer)

		n.logger.Debugf("'%s' peer is healthy!", peer.TcpAddress())
	} else {
		return nil
	}

	data, err := json.Marshal(JoinPeerRequest{From: n.metadata})
	if err != nil {
		return err
	}

	if _, err := peer.client.MakeCall(ctx, pb.Command_Sync, pb.Entity_Peer, data); err != nil {
		st, ok := status.FromError(err)
		if ok {
			if st.Code() == codes.AlreadyExists {
				n.logger.Debugw("Peer already known",
					"account", peer.Account, "addr", peer.TcpAddress())
				return nil
			}
		}
		n.logger.Errorf("Failed to join peer: %s", err)
		return err
	}

	return nil
}

func (n *Node) syncKnownPeers(ctx context.Context) error {
	for _, statusPeer := range n.knownPeers.peers {
		if !n.IsKnownPeer(statusPeer) {
			n.logger.Infof("Found new Peer %s", statusPeer.TcpAddress())

			if err := n.addDbPeer(statusPeer); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Node) syncVersion(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncPendingState(ctx context.Context, peer PeerNode) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncBlockHeaders(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncBlocks(ctx context.Context) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncTransactions(ctx context.Context, peer PeerNode) error {
	return fmt.Errorf("not implemented")
}

func (n *Node) syncAccountBalances(ctx context.Context, peer PeerNode) error {
	//balances, err := n.bc.

	return nil
}

func (n *Node) syncKeystore(ctx context.Context, peer PeerNode) error {
	return fmt.Errorf("not implemented")
}
