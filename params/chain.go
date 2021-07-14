package params

const (
	OpenDevNetworkId = "dev_rbn" // dev environment network id
	MainNetworkId    = "rbn"     // default main network id
)

const (
	TxPerBlockLimit int = 2560

	GenesisNetherLimit uint64 = 42e5 // Genesis block nether limit

	NetherLimit uint64 = 48e3 // Minimal nether fee limit may ever be.
	NetherPrice uint64 = 200  // Nether fee price multiplier per Coin
)

// tx fees
const (
	TxPrice         uint64 = 21e3 // Transaction cost modifier based on its value
	TxDataPrice     uint64 = 32e3 // Minimal transaction cost modifier based on data transfer amount
	NewAccountPrice uint64 = 24e3 // still not sure how i am supposed to use that, actually
	NetStoragePrice uint64 = 4196 // per data len/1024
)

// tx rewards
const (
	TxReward       uint64 = 64 // Reward multiplier per block transaction handled
	HardwareReward uint64 = 32 // Reward multiplier for network membership
)

// achievements
const (
	AchievementForYearOfPeering   = 42e7
	AchievementForE2SwapsVerified = 42e6 // note that this achievement works only for addressed cluster
)
