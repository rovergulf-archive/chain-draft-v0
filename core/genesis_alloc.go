package core

import "github.com/ethereum/go-ethereum/common"

type genesisAlloc map[common.Address]GenesisAccount

type GenesisAccount struct {
	Address common.Address `json:"address" yaml:"address"`
	Balance uint64         `json:"balance" yaml:"balance"`
	Auth    string         `json:"auth" yaml:"auth"`
}

func developerNetAlloc() genesisAlloc {
	return map[common.Address]GenesisAccount{
		common.HexToAddress("0x0000000000000000000000000000000000000000"): {
			Balance: 100e9,
		},
	}
}

func defaultMainNetAlloc() genesisAlloc {
	return map[common.Address]GenesisAccount{
		common.HexToAddress("0x2b2ee84325363350996487e45ebf487945a2547b"): {
			Balance: 100e9,
		},
		common.HexToAddress("0x2b2ee84325363350996487e45ebf487945a2547b"): {
			Balance: 100e9,
		},
		common.HexToAddress("0x2b2ee84325363350996487e45ebf487945a2547b"): {
			Balance: 100e9,
		},
	}
}
