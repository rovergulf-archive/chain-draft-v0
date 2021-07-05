package params

const (
	OpenDevNetworkId = "dev_rbn"
	MainNetworkId    = "rbn"
)

const (
	GenesisNetherLimit uint64 = 42e5

	NetherLimit uint64 = 4800 // Minimal nether limit may ever be.
	NetherPrice uint64 = 200  // Price per RNT

	TxPrice         uint64 = 21e3 //
	TxDataPrice     uint64 = 32e3
	NewAccountPrice uint64 = 24e3
	NetStoragePrice uint64 = 4196 // per 4kb

	TxReward uint64 = 0 // Reward fee per block transaction handled
)
