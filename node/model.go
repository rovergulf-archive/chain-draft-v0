package node

type StatusRes struct {
	LastHash   string `json:"block_hash,omitempty" yaml:"last_hash,omitempty"`
	Number     uint64 `json:"chain_length,omitempty" yaml:"chain_length,omitempty"`
	KnownPeers int    `json:"peers_known,omitempty" yaml:"known_peers,omitempty"`
	PendingTXs int    `json:"pending_txs,omitempty" yaml:"pending_t_xs,omitempty"`
}

type JoinPeerRequest struct {
	From PeerNode `json:"from" yaml:"from"`
}

type JoinPeerResult struct {
	Peers map[string]PeerNode `json:"known_peers" yaml:"known_peers"`
}

type VersionSyncReq struct {
	From PeerNode `json:"from" yaml:"from"`
}

type VersionSyncResult struct {
	From PeerNode `json:"from" yaml:"from"`
}

type BalanceSyncReq struct {
	From PeerNode `json:"from" yaml:"from"`
}

type BalanceSyncResult struct {
	From PeerNode `json:"from" yaml:"from"`
}

type TxAddRequest struct {
	Peer    PeerNode `json:"peer" yaml:"peer"`
	From    string   `json:"from" yaml:"from"`
	FromPwd string   `json:"from_pwd" yaml:"from_pwd"`
	To      string   `json:"to" yaml:"to"`
	Value   float64  `json:"value" yaml:"value"`
	Data    []byte   `json:"data" yaml:"data"`
}
