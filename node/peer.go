package node

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/rovergulf/rbn/client"
	"github.com/rovergulf/rbn/params"
	"github.com/spf13/viper"
	"strconv"
	"strings"
	"sync"
)

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

// PeerNode represents distributed node network metadata
type PeerNode struct {
	Id      string         `json:"id" yaml:"id"`
	Addrs   []string       `json:"addrs" yaml:"addrs"`
	Ip      string         `json:"ip" yaml:"ip"`
	Port    uint64         `json:"port" yaml:"port"`
	Root    bool           `json:"root" yaml:"root"`
	Account common.Address `json:"account" yaml:"account"`

	syncMode SyncMode

	// Whenever my node already established connection, sync with this Peer
	connected bool
	client    *client.NetherClient
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

func defaultPeer() PeerNode {
	return PeerNode{
		Ip:        DefaultNodeIP,
		Port:      DefaultNodePort,
		Root:      true,
		Account:   common.HexToAddress("0x3c0b3b41a1e027d3E759612Af08844f1cca0DdE3"),
		connected: false,
		syncMode:  SyncModeFull,
	}
}

// mainNetBootnodes returns the enode URLs of the P2P bootstrap nodes operated
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

func makeDefaultTrustedPeers() map[string]PeerNode {
	peers := make(map[string]PeerNode)
	for tcpAddr := range params.TreasurerAccounts {
		trustedNode := params.TreasurerAccounts[tcpAddr]
		addrParts := strings.Split(tcpAddr, ":")
		port, _ := strconv.ParseUint(addrParts[1], 10, 64)
		peers[tcpAddr] = NewPeerNode(addrParts[0], port, trustedNode, SyncModeFull)
	}
	return peers
}
