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

// StatusResult
type StatusResult struct {
	Head      string `json:"head" yaml:"head"`
	Genesis   string `json:"genesis"  yaml:"genesis"`
	NetworkId string `json:"network_id" yaml:"network_id"`
	Uptime    int64  `json:"uptime" yaml:"uptime"`
}

type PeerInfo struct {
	Id        string         `json:"id" yaml:"id"`
	Enode     string         `json:"enode" yaml:"enode"`
	Name      string         `json:"name" yaml:"name"`
	Protocols map[string]int `json:"protocols" yaml:"protocols"`
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
