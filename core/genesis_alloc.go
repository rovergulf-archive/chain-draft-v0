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
		common.HexToAddress("0x10dc3b9e09bc819b9f6f4def14fdb879c4ab0c7d"): {
			Balance: 110e9,
		},
		common.HexToAddress("0x36527b4481018dff6d3400a2271d070910453420"): {
			Balance: 110e9,
		},
		common.HexToAddress("0x3c0b3b41a1e027d3e759612af08844f1cca0dde3"): {
			Balance: 110e9,
		},
		common.HexToAddress("0x40b2121f4eb40B6863A08D08C567CC1C995f971F"): {
			Balance: 110e6,
		},
	}
}
