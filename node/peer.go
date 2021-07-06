package node

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"sync"
)

var (
	peerPrefix = []byte("peers/")
)

type knownPeers struct {
	peers map[string]PeerNode
	lock  *sync.RWMutex
}

type SyncMode string

func (sm SyncMode) String() string {
	return string(sm)
}

const (
	SyncModeDefault SyncMode = "default" // only block headers
	SyncModeAccount SyncMode = "account" // download node account related transactions and blocks
	SyncModeFull    SyncMode = "full"    // sync full chain
)

// PeerNode represents distributed node network metadata
type PeerNode struct {
	Ip      string         `json:"ip" yaml:"ip"`
	Port    uint64         `json:"port" yaml:"port"`
	Root    bool           `json:"root" yaml:"root"`
	Account common.Address `json:"account" yaml:"account"`

	syncMode SyncMode

	// Whenever my node already established connection, sync with this Peer
	connected bool
	client    *Client
}

func NewPeerNode(ip string, port uint64, address common.Address, mode SyncMode) PeerNode {
	return PeerNode{
		Ip:       ip,
		Port:     port,
		Root:     ip == DefaultNodeIP && port == DefaultNodePort,
		Account:  address,
		syncMode: mode,
	}
}

func (pn *PeerNode) SyncMode() string {
	return pn.syncMode.String()
}

// TcpAddress returns tcp node address
func (pn *PeerNode) TcpAddress() string {
	return fmt.Sprintf("%s:%d", pn.Ip, pn.Port)
}

// ApiProtocol returns http protocol
func (pn *PeerNode) ApiProtocol() string {
	if pn.Port == HttpSSLPort {
		return "https"
	}

	return "http"
}

// ApiAddress returns HTTP server listen address
func (pn *PeerNode) ApiAddress() string {
	return fmt.Sprintf("%s:%s", viper.GetString("http.addr"), viper.GetString("http.port"))
}

// HttpApiAddress returns full API server URL
func (pn *PeerNode) HttpApiAddress() string {
	return fmt.Sprintf("%s://%s", pn.ApiProtocol(), pn.ApiAddress())
}

// addPeer saves new peer to node storage
func (n *Node) addPeer(peer PeerNode) error {
	n.logger.Info("n.addPeer", peer)

	pn, err := peer.Serialize()
	if err != nil {
		return err
	}

	n.knownPeers.lock.Lock()
	defer n.knownPeers.lock.Unlock()
	return n.db.Update(func(txn *badger.Txn) error {
		key := append(peerPrefix, peer.Account.Bytes()...)
		if err := txn.Set(key, pn); err != nil {
			return err
		} else {
			n.knownPeers.peers[peer.Account.String()] = peer
		}
		return nil
	})
}

// removePeer deletes peer from node storage
func (n *Node) removePeer(peer PeerNode) error {
	n.knownPeers.lock.Lock()
	defer n.knownPeers.lock.Unlock()
	return n.db.Update(func(txn *badger.Txn) error {
		key := append(peerPrefix, peer.Account.Bytes()...)
		if err := txn.Delete(key); err != nil {
			return err
		} else {
			delete(n.knownPeers.peers, peer.Account.String())
		}

		return nil
	})
}

func (pn *PeerNode) Serialize() ([]byte, error) {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	if err := enc.Encode(pn); err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

func (pn *PeerNode) Deserialize(src []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(src))
	return decoder.Decode(pn)
}

func collectPeerUrls(nodes map[string]PeerNode) []string {
	var peers []string

	for peer := range nodes {
		node := nodes[peer]
		peers = append(peers, node.TcpAddress())
	}

	return peers
}
