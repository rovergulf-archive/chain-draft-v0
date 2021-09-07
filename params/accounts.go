package params

import "github.com/ethereum/go-ethereum/common"

var DevTreasurerAccounts = map[string]common.Address{
	"": common.HexToAddress("TODO"),
}

// TreasurerAccounts is the coinbase and nether pool network trusted accounts
// which would distribute awards for network membership and etc
var TreasurerAccounts = map[string]common.Address{
	DefaultNodeAddr: common.HexToAddress("0x3c0b3b41a1e027d3E759612Af08844f1cca0DdE3"),
}
