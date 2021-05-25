# Changelog
All notable changes to this project will be documented in this file.


## [Unreleased] v0.1.0

## 25 May 2021

### Added
- CLI backup and balances commands templates
- prepared Node and PeerNode structs to handle network

### Changed
- config package containing common options

### Fixed

### Removed


## 24 May 2021

### Changed
- Updated CLI usage
- badger db setup moved to separate package


## 23 May 2021

### Added
- bitcoin base58 encoded wallet address support
- TODO: change to [etherium based address](https://pkg.go.dev/github.com/ethereum/go-ethereum/crypto/secp256k1) using `secp256k1`
- node package to handle blockchain peers

### Fixed
- blockchain init/continue handlers


## 22 May 2021

### Added
- UTXO set
- accounts package to handle wallet addresses
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

