package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/rovergulf/rbn/core"
	"io"
	"io/ioutil"
	"log"
	"net"
)

const (
	protocol      = "tcp"
	version       = 1
	commandLength = 12
)

var (
	nodeAddress     string
	mineAddress     string
	KnownNodes      = []string{"localhost:9420"}
	blocksInTransit = [][]byte{}
	memoryPool      = make(map[string]core.Transaction)
)

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

func CmdToBytes(cmd string) []byte {
	var bytes [commandLength]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

func RequestBlocks() error {
	for _, node := range KnownNodes {
		if err := SendGetBlocks(node); err != nil {
			return err
		}
	}

	return nil
}

func SendAddr(address string) {
	nodes := Addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(CmdToBytes("addr"), payload...)

	SendData(address, request)
}

func SendBlock(addr string, b *core.Block) error {
	serializedBlock, err := b.Serialize()
	if err != nil {
		return err
	}
	data := Block{
		AddrFrom: nodeAddress,
		Block:    serializedBlock,
	}

	payload := GobEncode(data)
	request := append(CmdToBytes("block"), payload...)

	SendData(addr, request)
	return nil
}

func SendData(addr string, data []byte) error {
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return err
	}
	defer conn.Close()

	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}

func SendInv(address, kind string, items [][]byte) error {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)

	return SendData(address, request)
}

func SendGetBlocks(address string) error {
	payload := GobEncode(GetBlocks{nodeAddress})
	request := append(CmdToBytes("getblocks"), payload...)

	return SendData(address, request)
}

func SendGetData(address, kind string, id []byte) error {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CmdToBytes("getdata"), payload...)

	return SendData(address, request)
}

func SendTx(addr string, tnx *core.Transaction) error {
	serializedTx, err := tnx.Serialize()
	if err != nil {
		return err
	}

	data := Tx{nodeAddress, serializedTx}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	return SendData(addr, request)
}

func SendVersion(addr string, chain *core.Blockchain) error {
	bestHeight, err := chain.GetBestHeight()
	if err != nil {
		return err
	}

	payload := GobEncode(Version{version, bestHeight, nodeAddress})

	request := append(CmdToBytes("version"), payload...)

	return SendData(addr, request)
}

func HandleAddr(request []byte) error {
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	fmt.Printf("there are %d known nodes\n", len(KnownNodes))
	return RequestBlocks()
}

func HandleBlock(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block, err := core.DeserializeBlock(blockData)
	if err != nil {
		return err
	}

	fmt.Println("Recevied a new block!")
	if err := chain.AddBlock(block); err != nil {
		return err
	}

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		if err := SendGetData(payload.AddrFrom, "block", blockHash); err != nil {
			return err
		}

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := core.UTXOSet{chain}
		return UTXOSet.Reindex()
	}

	return nil
}

func HandleInv(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload Inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		if err := SendGetData(payload.AddrFrom, "block", blockHash); err != nil {
			return err
		}

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			return SendGetData(payload.AddrFrom, "tx", txID)
		}
	}

	return nil
}

func HandleGetBlocks(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload GetBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks, err := chain.GetBlockHashes()
	if err != nil {
		return err
	}

	return SendInv(payload.AddrFrom, "block", blocks)
}

func HandleGetData(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload GetData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return err
		}

		return SendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]

		return SendTx(payload.AddrFrom, &tx)
	}

	return nil
}

func HandleTx(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload Tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	txData := payload.Transaction
	tx, err := core.DeserializeTransaction(txData)
	if err != nil {
		return err
	}
	memoryPool[hex.EncodeToString(tx.ID)] = *tx

	fmt.Printf("%s, %d", nodeAddress, len(memoryPool))

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				if err := SendInv(node, "tx", [][]byte{tx.ID}); err != nil {
					return err
				}
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(mineAddress) > 0 {
			MineTx(chain)
		}
	}

	return nil
}

func MineTx(chain *core.Blockchain) error {
	var txs []*core.Transaction

	for id := range memoryPool {
		fmt.Printf("tx: %s\n", memoryPool[id].ID)
		tx := memoryPool[id]
		if err := chain.VerifyTransaction(&tx); err != nil {
			return err
		}
		txs = append(txs, &tx)
	}

	if len(txs) == 0 {
		return fmt.Errorf("all transactions are invalid")
	}

	cbTx := core.CoinbaseTx(mineAddress, "")
	txs = append(txs, cbTx)

	newBlock, err := chain.MineBlock(txs)
	if err != nil {
		return err
	}

	UTXOSet := core.UTXOSet{chain}
	if err := UTXOSet.Reindex(); err != nil {
		return err
	}

	fmt.Println("New Block mined")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range KnownNodes {
		if node != nodeAddress {
			if err := SendInv(node, "block", [][]byte{newBlock.Hash}); err != nil {
				return err
			}
		}
	}

	if len(memoryPool) > 0 {
		if err := MineTx(chain); err != nil {
			return err
		}
	}

	return nil
}

func HandleVersion(request []byte, chain *core.Blockchain) error {
	var buff bytes.Buffer
	var payload Version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		return err
	}

	bestHeight, err := chain.GetBestHeight()
	if err != nil {
		return err
	}
	otherHeight := payload.BestHeight

	if bestHeight < otherHeight {
		if err := SendGetBlocks(payload.AddrFrom); err != nil {
			return err
		}
	} else if bestHeight > otherHeight {
		if err := SendVersion(payload.AddrFrom, chain); err != nil {
			return err
		}
	}

	if !NodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}

	return nil
}

func HandleConnection(conn net.Conn, chain *core.Blockchain) error {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}
	command := BytesToCmd(req[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		return HandleAddr(req)
	case "block":
		return HandleBlock(req, chain)
	case "inv":
		return HandleInv(req, chain)
	case "getblocks":
		return HandleGetBlocks(req, chain)
	case "getdata":
		return HandleGetData(req, chain)
	case "tx":
		return HandleTx(req, chain)
	case "version":
		return HandleVersion(req, chain)
	default:
		return fmt.Errorf("unknown command")
	}
}

func StartServer(nodeID, minerAddress string) error {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	mineAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	chain, err := core.ContinueBlockchain(core.Options{
		DbFilePath: "",
		Address:    "",
		NodeId:     nodeID,
		Badger:     badger.Options{},
		Logger:     nil,
		Tracer:     nil,
	})
	if err != nil {

	}
	defer chain.Shutdown()

	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// log err
		}
		go func() {
			if err := HandleConnection(conn, chain); err != nil {
				// log err
			}
		}()

	}
}

func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func NodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}
