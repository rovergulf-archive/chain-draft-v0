package node

type CallRequest struct {
	Code uint64 `json:"code" yaml:"code"`
	Data []byte `json:"data" yaml:"data"`
	From string `json:"from" yaml:"from"` //
}

// CallResult represents standard
type CallResult struct {
	Code uint64 `json:"code" yaml:"code"`
	Data []byte `json:"data" yaml:"data"`
}

type StatusResult struct {
	Head      string `json:"head" yaml:"head"`
	Genesis   string `json:"genesis"  yaml:"genesis"`
	NetworkId string `json:"network_id" yaml:"network_id"`
	Uptime    int64  `json:"uptime" yaml:"uptime"`
}

type JoinPeerRequest struct {
}

type JoinPeerResult struct {
	Peers map[string]PeerNode `json:"known_peers" yaml:"known_peers"`
}

type VersionSyncReq struct {
}

type VersionSyncResult struct {
}

type BalanceSyncReq struct {
}

type BalanceSyncResult struct {
}

type TxAddRequest struct {
	From    string  `json:"from" yaml:"from"`
	FromPwd string  `json:"from_pwd" yaml:"from_pwd"`
	To      string  `json:"to" yaml:"to"`
	Value   float64 `json:"value" yaml:"value"`
	Data    []byte  `json:"data" yaml:"data"`
}
