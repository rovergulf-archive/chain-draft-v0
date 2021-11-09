package core

import (
	"bytes"
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
	"reflect"
	"regexp"
	"strconv"
)

const (
	DbFileName = "chain.db"
)

var (
	emptyHash = common.HexToHash("")
)

var (
	ErrGenesisNotExists     = errors.New("genesis does not exists")
	ErrBalanceNotExists     = errors.New("balance does not exists")
	ErrBalanceAlreadyExists = errors.New("balance already exists")
	ErrBlockNotExists       = errors.New("block does not exists")
	ErrBlockAlreadyExists   = errors.New("block already exists")
	ErrTxNotExists          = errors.New("transaction does not exists")
	ErrTxAlreadyExists      = errors.New("transaction already exists")
	ErrInvalidRewardData    = errors.New("invalid reward tx data")
	ErrReceiptNotExists     = errors.New("receipt does not exists")
	ErrReceiptAlreadyExists = errors.New("receipt already exists")
)

var (
	lastHashKey        = []byte("lh")
	genesisKey         = []byte("gen")
	genesisBlockKey    = []byte("root")
	blocksPrefix       = []byte("blocks/")
	blockNumbersPrefix = []byte("blockNums/")
	blockHeadersPrefix = []byte("headers/")
	balancesPrefix     = []byte("balances/")
	txsPrefix          = []byte("txs/")
	receiptsPrefix     = []byte("receipts/")
)

func blockDbPrefix(hash common.Hash) []byte {
	return append(blocksPrefix, hash.Bytes()...)
}

func blockNumDbPrefix(number uint64) []byte {
	numStr := strconv.FormatUint(number, 10)
	prefix := []byte(numStr)
	return append(blockNumbersPrefix, prefix...)
}

func blockHeaderDbPrefix(hash common.Hash) []byte {
	return append(blockHeadersPrefix, hash.Bytes()...)
}

func balanceDbPrefix(addr common.Address) []byte {
	return append(balancesPrefix, addr.Bytes()...)
}

func txDbPrefix(hash common.Hash) []byte {
	return append(txsPrefix, hash.Bytes()...)
}

func receiptDbPrefix(hash common.Hash) []byte {
	return append(receiptsPrefix, hash.Bytes()...)
}

func IsHashEmpty(hash common.Hash) bool {
	return bytes.Compare(hash.Bytes(), emptyHash.Bytes()) == 0
}

// IsValidAddress validate hex address
func IsValidAddress(iaddress interface{}) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	switch v := iaddress.(type) {
	case string:
		return re.MatchString(v)
	case common.Address:
		return re.MatchString(v.Hex())
	default:
		return false
	}
}

// IsZeroAddress validate if it's a 0 address
func IsZeroAddress(iaddress interface{}) bool {
	var address common.Address
	switch v := iaddress.(type) {
	case string:
		address = common.HexToAddress(v)
	case common.Address:
		address = v
	default:
		return false
	}

	zeroAddressBytes := common.FromHex("0x0000000000000000000000000000000000000000")
	addressBytes := address.Bytes()
	return reflect.DeepEqual(addressBytes, zeroAddressBytes)
}

// CalcGasCost calculate gas cost given gas limit (units) and gas price (wei)
func CalcGasCost(gasLimit uint64, gasPrice *big.Int) *big.Int {
	gasLimitBig := big.NewInt(int64(gasLimit))
	return gasLimitBig.Mul(gasLimitBig, gasPrice)
}

// SigRSV signatures R S V returned as arrays
func SigRSV(isig interface{}) ([32]byte, [32]byte, uint8) {
	var sig []byte
	switch v := isig.(type) {
	case []byte:
		sig = v
	case string:
		sig, _ = hexutil.Decode(v)
	}

	sigstr := common.Bytes2Hex(sig)
	rS := sigstr[0:64]
	sS := sigstr[64:128]
	R := [32]byte{}
	S := [32]byte{}
	copy(R[:], common.FromHex(rS))
	copy(S[:], common.FromHex(sS))
	vStr := sigstr[128:130]
	vI, _ := strconv.Atoi(vStr)
	V := uint8(vI + 27)

	return R, S, V
}
