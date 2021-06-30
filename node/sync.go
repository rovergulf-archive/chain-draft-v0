package node

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rovergulf/rbn/core"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

func (n *Node) sync(ctx context.Context) error {
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
	for _, peer := range n.knownPeers {
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

		//if err := n.syncBlocks(peer); err != nil {
		//	n.logger.Error("Unable to sync blocks: ", err)
		//	continue
		//}

		//if err := n.syncKnownPeers(); err != nil {
		//	n.logger.Error("Unable to sync knonw peers", err)
		//	continue
		//}

		//if err := n.syncPendingTXs(peer, status.PendingTXs); err != nil {
		//	n.logger.Error(err)
		//	continue
		//}
	}
}

func (n *Node) syncBlocks(peer PeerNode, status *StatusRes) error {
	localBlockNumber := n.bc.ChainLength.Uint64()

	// If the peer has no blocks, ignore it
	if status.LastHash == "" {
		return nil
	}

	// If the peer has less blocks than us, ignore it
	if status.Number < localBlockNumber {
		return nil
	}

	// If it's the genesis block and we already synced it, ignore it
	if status.Number == 0 && len(n.bc.LastHash.Bytes()) == 0 {
		return nil
	}

	// Display found 1 new block if we sync the genesis block 0
	newBlocksCount := status.Number - localBlockNumber
	if localBlockNumber == 0 && status.Number == 0 {
		newBlocksCount = 1
	}
	n.logger.Infof("Found %d new blocks from Peer %s", newBlocksCount, peer.TcpAddress())

	blocks, err := n.fetchBlocksFromPeer(peer, n.bc.LastHash)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if err := n.bc.AddBlock(block); err != nil {
			return err
		}

		n.newSyncBlocks <- *block
	}

	return nil
}

func (n *Node) syncKnownPeers(status *StatusRes) error {
	for _, statusPeer := range status.KnownPeers {
		if !n.IsKnownPeer(statusPeer) {
			fmt.Printf("Found new Peer %s\n", statusPeer.TcpAddress())

			if err := n.addPeer(statusPeer); err != nil {
				return err
			}
		}
	}

	return nil
}

//func (n *Node) syncPendingTXs(peer PeerNode, txs []*core.Transaction) error {
//	for _, tx := range txs {
//		err := n.AddPendingTX(tx, peer)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

func (n *Node) joinKnownPeer(peer PeerNode) error {
	if peer.connected {
		return nil
	}

	url := fmt.Sprintf(
		"%s://%s%s?%s=%s&%s=%d",
		peer.ApiProtocol(),
		peer.TcpAddress(),
		endpointAddPeer,
		endpointAddPeerQueryKeyIP,
		n.metadata.Ip,
		endpointAddPeerQueryKeyPort,
		n.metadata.Port,
	)

	res, err := http.Get(url)
	if err != nil {
		return err
	}

	var addPeerRes AddPeerRes
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&addPeerRes); err != nil {
		return err
	}

	if addPeerRes.Error != "" {
		return fmt.Errorf(addPeerRes.Error)
	}

	knownPeer := n.knownPeers[peer.TcpAddress()]
	knownPeer.connected = addPeerRes.Success

	if err := n.addPeer(knownPeer); err != nil {
		return err
	}

	if !addPeerRes.Success {
		return fmt.Errorf("unable to join KnownPeers of '%s'", peer.TcpAddress())
	}

	return nil
}

func (n *Node) fetchBlocksFromPeer(peer PeerNode, fromBlock common.Hash) ([]*core.Block, error) {
	n.logger.Infof("Importing blocks from Peer %s...\n", peer.TcpAddress())

	url := fmt.Sprintf(
		"%s://%s%s?%s=%s",
		peer.ApiProtocol(),
		peer.TcpAddress(),
		endpointSync,
		endpointSyncQueryKeyFromBlock,
		fmt.Sprintf("%x", fromBlock),
	)

	res, err := http.Get(url)
	if err != nil {
		n.logger.Errorf("Unable to decode sync res: %s", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(res.Status)
	}

	var syncRes SyncRes
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&syncRes); err != nil {
		n.logger.Errorf("Unable to decode sync res: %s", err)
		return nil, err
	}

	return syncRes.Blocks, nil
}
