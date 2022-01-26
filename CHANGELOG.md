# Changelog
All notable changes to this project will be documented in this file.


## [Unreleased] v0.1.0

## 27 Jan 2022

### Added

### Changed

### Fixed

### Removed


## 26 Jan 2022

### Added

### Changed
- restructurize
- update dependencies

### Fixed

### Removed


## 13 Nov 2021

### Added
- chain utils can parse private key string

### Changed
- Dependabot to weekly frequency checks
- Prepare dgraph driver
- Prepare API Handler interface


## 10 Sep 2021

### Changed
- Updated uber zap logger dependency version


## 9 Sep 2021

### Added
- Basic [dockerfile](Dockerfile)


## 8 Sep 2021

### Added
- GitHub actions Go fmt and vet checks (tests would be added as better design appears)
- new p2p handling templates based on Etherium `p2p` 

### Removed
- gRPC and protobuf usage


## 6 Sep 2021

### Changed
- `proto` package renamed and moved to `node/pb`. `scripts/genproto.sh` fixed as well

### Fixed
- Downgrade go-multiaddr and libp2p-go-core libs, due runtime error

### Removed
- libp2p usage, there is etherium library, simply working right up here. What did I even tried to do?
- legacy `etcd/raft` import


## 5 Sep 2021

### Added
- basic p2p host usage and connection

### Changed

### Fixed
- close topic, dht and subs on graceful shutdown

### Removed


## 4 Sep 2021

### Added
- libp2p usage

### Changed
- pb files regenerated without gRPC server, probably should be removed

### Removed
- grpc server usage


## 14 Jul 2021

### Added
- Reward transactions handle

## 13 Jul 2021

### Added
- `traceutil` package parent context option wrapper func


## 9 Jul 2021

### Added
- sync known peers (to be tested)
- prefixes for all kv database keys in BlockChain, also added wrapper functions
- wallets manager address existsing method
- wallet lock status method which returns "Un/Locked" string


## 7 Jul 2021

### Added
- Added [Contributing guide](CONTRIBUTING.md)
- I have lost a lot of time at trying to implement common database interface,
  only after lot of spent time I get that it is not even good fits the idea
  of separate interfaces and storages. It would be described in further docs:
  this chain would be as disturbed, as you do not need to have the whole chain backup. 
  Only the data compared with your node and accounts registered at network.


## 6 Jul 2021

### Added
- Use receipts as transaction applied return result

### Changed
- node blockchain state run does not load genesis now
- peer removes itself from mem cache discovery
- gRPC client moved to separate `client` package – probably would be renamed and/or moved
- Denomination for RNT renamed to Coin and would be used as Rovergulf Coins

### Fixed

### Removed


## 5 Jul 2021

### Added
- Transaction fee calculation

### Changed
- Block and Balance structs moved to `core/types` package


## 4 Jul 2021

### Added
- more `params` package constants, to be used in network consensus verification
  and other places where constant values are important

### Changed
- Chain Genesis initialization now have hardcoded default values
- `core` logic now separated from its types, as it would represent database interface of blockchain
- some of `core` package used types now moved to `types` subdirectory package
- No etherium **Gas** naming, use **Nether** as the lowest denomination of **Rovergulf Native Token**
- `address` flag usage for cli application would be optimized

### Removed
- `node-id` flag usage


## 3 Jul 2021

### Added
- node account dump command
- transaction signer verification

### Changed
- Genesis initialization

### Removed
- legacy sync methods


## 1 Jul 2021

### Changed
- provide address on node run, to switch node used account

### Fixed
- balance database saving

### Removed
- PoW usage


## 30 Jun 2021

### Added
- node stop CLI command

### Changed
- get account key handler
- sign transactions with open wallet
- node currently asks for account passphrase to run (behavior should be updated)

### Fixed
- Genesis db encoding - removed `Transaction.MarshalJSON` method


## 29 Jun 2021

### Added
- Added get block http handler
- gRPC node connection interface instead custom TCP
- genproto.sh script to generate rpc command
- some gRPC handlers
- change account auth phrase

### Changed
- wallets creation use mnemonic passphrase by `--mnemonic` 
  flag value now, which is true by default.


## 28 Jun 2021

### Added
- Genesis initialization from file

### Changed
- Updated transaction signing using etherium SDK
- Use Ether-like wallet balance design instead Bitcoin UTXO


## 3 Jun 2021

### Changed
- `Wallets` now are `Manager`
- `repo` package moved to `database/badgerdb` 
  to prepare multiple database backend interface
  - [Dgraph Go client](https://github.com/dgraph-io/dgo)


## 1 Jun 2021

### Added
- Prepare genesis from file
- `tx get --id $TRANSACTION_ID` command


## 31 May 2021

### Added
- Prepared node multiple network interfaces
- etherium-go SDK wallet `common.Address` support

### Changed
- `accounts` package moved to `wallets`

### Fixed
- Send command public key length
- Get wallet balance
- List balances command output


## 25 May 2021

### Added
- CLI backup and balances commands templates
- Prepared Node and PeerNode structs to handle network

### Changed
- `config` package containing common options


## 24 May 2021

### Changed
- Updated CLI usage
- Badger DB setup moved to separate package


## 23 May 2021

### Added
- Bitcoin base58 encoded wallet address support
- TODO: change to [etherium based address](https://pkg.go.dev/github.com/ethereum/go-ethereum/crypto/secp256k1) using `secp256k1`
- `node` package to handle blockchain peers

### Fixed
- blockchain init/continue handlers


## 22 May 2021

### Added
- UTXO set
- `accounts` package to handle wallet addresses
- base58 encoding (temporarily to supply bitcoin like address)

### Changed
- updated transactions


## 21 May 2021

### Changed
- use [badger db](https://github.com/dgraph-io/badger) instead bbolt


## 20 May 2021

### Added
- core transactions handling


## 19 May 2021

### Added
- Basic blockchain prototype
- [bbolt](https://github.com/etcd-io/bbolt) db storage


[Unreleased]: https://github.com/rovergulf/engine/v0.1.0...main
[v0.2.0]: https://github.com/rovergulf/engine/compare/v0.1.0...v0.2.0
[v0.0.1]: https://github.com/rovergulf/engine/tree/v0.1.0

