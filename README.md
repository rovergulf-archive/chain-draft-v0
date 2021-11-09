![GitHub](https://img.shields.io/github/license/rovergulf/rbn)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/rovergulf/rbn)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/rovergulf/rbn)

## Rovergulf BlockChain Network

[RBN](https://chain.rovergulf.net) - BlockChain algorithm implemented by [Rovergulf Engineers](https://rovergulf.net) team based on [Etherium SDK](https://github.com/ethereum/go-ethereum)

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
git clone github.com/rovergulf/chain
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


### List keystore accounts
If you run node without using wallets manager, it will automatically allocate a new one, place it to keystore,  
and you will be able to import that account passphrase using node admin API
```shell
rbn wallets list 
```

## Coins

### Rovergulf Coins

An extra reward for platform achievements and new block validations

### Consensus

Read more about Rovergulf Smart Chain algorithm in [documentation](https://chain.rovergulf.net/docs/nether)  
Short representation of it can be found in [consensus/README.md](consensus/README.md)

---

## Contributing

Read [Contributing guide](CONTRIBUTING.md)


## Maintainers

**Rovergulf Engineering Team** <team@rovergulf.net>  

Author: Dmitrii Limonov <d@rovergulf.net>  
