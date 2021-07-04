# Changelog
All notable changes to this project will be documented in this file.


## [Unreleased] v0.1.0


## 3 Jul 2021

### Added
- more `params` package constants, to be used in network consensus verification

### Changed
- Chain genesis initialization now have hardcoded default values 

### Fixed

### Removed


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

