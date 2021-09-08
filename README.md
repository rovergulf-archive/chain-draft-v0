![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rovergulf/rbn)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/rovergulf/rbn)

## Rovergulf BlockChain Network

[RBN](https://chain.rovergulf.net) - BlockChain algorithm implemented by Rovergulf Engineers team based on [Etherium SDK](https://github.com/ethereum/go-ethereum)

### Information
- Documentation: [RBN Docs](https://chain.rovergulf.net/docs)
- BlockChain network available at `swarm.rovergulf.net`

---
## Development

### Simple source build
```shell
go build -o rbn cmd/cli/main.go

./rbn help

# or just

go run cmd/cli/main.go help
```


## Installation

### From source
```shell
git clone github.com/rovergulf/rbn
wget dl.rovergulf.net/rbn:$RELEASE # TBD platform support 
```

### Containers
```shell
TBD
```

### Verify installation
```shell
# validate using
rbn help
# or
rbn --version
```

---

## Run node
```shell
rbn node run

# to get more opts
rbn node run --help

# TBD: sync modes, etc
```

## Manage accounts
```shell
# list node accounts
# if you run node without your eth key,
rbn wallets list
```

## Coins

### Rovergulf Coins

An extra reward for platform achievements (for a [swap](https://swap.rovergulf.net) platform) and new block validations

---

## Contributing

Read [Contributing guide](CONTRIBUTING.md)


## Maintainers

**Rovergulf Engineering Team** <team@rovergulf.net>  

Author: Dmitrii Limonov <d@rovergulf.net>  
