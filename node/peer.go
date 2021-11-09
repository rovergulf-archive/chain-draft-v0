package node

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/rovergulf/chain/params"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"sync"
)

type Peer struct {
	id      string
	version string

	peer *p2p.Peer
	rw   p2p.MsgReadWriter

	logger *zap.SugaredLogger
}

func NewPeer(peer *p2p.Peer, rw p2p.MsgReadWriter) *Peer {
	p := &Peer{
		id:   peer.ID().String(),
		peer: peer,
		rw:   rw,
	}

	return p
}

func (p *Peer) Close() {
	p.peer.Disconnect(p2p.DiscQuitting)
}

func (p *Peer) ID() string {
	return p.id
}

func (p *Peer) Version() string {
	return p.version
}

func (p *Peer) Info() {

}

var (
	peerPrefix = []byte("peers/")
)

func peerDbPrefix() []byte {
	return append(peerPrefix)
}

type knownPeers struct {
	peers map[string]PeerNode
	lock  *sync.RWMutex
}

func (k knownPeers) Exists(addr string) bool {
	k.lock.RLock()
	_, ok := k.peers[addr]
	k.lock.RUnlock()
	return ok
}

func (k knownPeers) GetPeers() map[string]PeerNode {
	var peers map[string]PeerNode
	k.lock.RLock()
	peers = k.peers
	k.lock.RUnlock()
	return peers
}

func (k knownPeers) GetPeer(addr string) (PeerNode, bool) {
	var peer PeerNode
	var ok bool
	k.lock.RLock()
	peer, ok = k.peers[addr]
	k.lock.RUnlock()
	return peer, ok
}

func (k knownPeers) AddPeer(addr string, peer PeerNode) {
	k.lock.Lock()
	k.peers[addr] = peer
	k.lock.Unlock()
}

func (k knownPeers) DeletePeer(addr string) {
	k.lock.Lock()
	delete(k.peers, addr)
	k.lock.Unlock()
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

// PeerNode represents distributed peer-node network metadata
type PeerNode struct {
	account common.Address

	syncMode SyncMode

	// Whenever my node already established connection, sync with this Peer
	connected bool

	id string

	peer *p2p.Peer
	rw   p2p.MsgReadWriter
}

func NewPeerNode(peer *p2p.Peer, rw p2p.MsgReadWriter) PeerNode {
	p := PeerNode{
		id:   peer.ID().String(),
		peer: peer,
		rw:   rw,
	}

	return p
}

func (pn *PeerNode) SyncMode() string {
	return pn.syncMode.String()
}

// TcpAddress returns tcp node address
func (pn *PeerNode) TcpAddress() string {
	return pn.peer.RemoteAddr().String()
}

// RemoteAddress returns peer remote url
func (pn *PeerNode) RemoteAddress() string {
	return pn.peer.RemoteAddr().String()
}

// ApiProtocol returns http protocol
func (pn *PeerNode) ApiProtocol() string {
	if viper.GetInt("http.port") == HttpSSLPort {
		return "https"
	}

	return "http"
}

// ApiAddress returns HTTP server listen address
func (pn *PeerNode) ApiAddress() string {
	return fmt.Sprintf("%s:%s", viper.GetString("http.addr"), viper.GetString("http.port"))
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

// mainNetBootNodes returns the enode URLs of the P2P bootstrap nodes operated
// by the Rovergulf Engineers running the V5 discovery protocol.
func mainNetBootNodes() []*enode.Node {
	nodes := make([]*enode.Node, len(params.MainNetBootNodes))
	for i, url := range params.MainNetBootNodes {
		var err error
		nodes[i], err = enode.Parse(enode.ValidSchemes, url)
		if err != nil {
			panic("invalid node URL: " + err.Error())
		}
	}
	return nodes
}
