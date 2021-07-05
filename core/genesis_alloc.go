package core

import "github.com/ethereum/go-ethereum/common"

type genesisAlloc map[common.Address]GenesisAccount

type GenesisAccount struct {
	Address common.Address `json:"address" yaml:"address"`
	Balance uint64         `json:"balance" yaml:"balance"`
}

func developerNetAlloc() genesisAlloc {
	return map[common.Address]GenesisAccount{
		common.HexToAddress("0x0000000000000000000000000000000000000000"): {
			Balance: 1e12,
		},
		common.HexToAddress("0x0000000000000000000000000000000000000000"): {
			Balance: 1e9,
		},
		common.HexToAddress("0x0000000000000000000000000000000000000000"): {
			Balance: 1e6,
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
		common.HexToAddress("0xF5f998c761F0CE7e2b15df323e6862D0C31c9F6F"): {
			Balance: 110e9,
		},
		common.HexToAddress("0x40b2121f4eb40B6863A08D08C567CC1C995f971F"): {
			Balance: 110e9,
		},
	}
}
