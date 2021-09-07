package params

const (
	DevNodeAddr = "" // Frankfurt based
)

const (
	DefaultNodeAddr = "127.0.0.1:9420" // Frankfurt based

	TokyoNodeAddr      = ""
	SydneyNodeAddr     = ""
	MontrealNodeAddr   = ""
	SaoPaoloNodeAddr   = ""
	LosAngelesNodeAddr = ""
)

// MainNetBootNodes are the enode URLs of the P2P bootstrap nodes running on the Rovergulf Blockchain network.
var MainNetBootNodes = []string{
	"enode://f6f6d63943a26beb48695c141e10bfb7277c448e3b401c40e0b37fb42e791ab72afc08d3bcefdd55d2b0631b1697308e164565496edaaa4e50149d975e3ff909@127.0.0.1:9420",
	"enode://8a83023555d2cbadf5c8f34b77fe6687fce576b7747241f17eced939ab713a00039ed605dc53bce1b8ece741c5cc509741d7963eee097d0fb06847f978577c09@127.0.0.1:9421",
}

var MainNetV5BootNodes = []string{
	"enr:-Iu4QDJ189YKzPMbg_O-Sct9tZyQ1akttoBoq2Sn_tWwM9NPHuxq0AofvyZeb29bPAZiN2ZPYrX_TA0FUFVpio1jraoBgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQP29tY5Q6Jr60hpXBQeEL-3J3xEjjtAHEDgs3-0Lnkat4N0Y3CCJMyDdWRwgiTM",
	"enr:-Iu4QHuwLM2S-U7VeNsHXLj7Gu9hk-6Kjn9RMcn8FQSsnQODR6gMKJ4Incbxe-9r1d3AWztQNlPtZ8TLsTWdIGo7AW8BgmlkgnY0gmlwhH8AAAGJc2VjcDI1NmsxoQOKgwI1VdLLrfXI80t3_maH_OV2t3RyQfF-ztk5q3E6AIN0Y3CCJM2DdWRwgiTN",
}
