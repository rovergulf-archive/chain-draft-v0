package node

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/rovergulf/rbn/core"
	"io"
	"log"
	"net"
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

func (n *Node) SendTx(addr string, tnx *core.Transaction) error {
	serializedTx, err := tnx.Serialize()
	if err != nil {
		return err
	}

	data := Tx{
		AddrFrom:    addr,
		Transaction: serializedTx,
	}

	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	return n.SendData(addr, request)
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

func (n *Node) SendData(addr string, data []byte) error {
	const protocol = "tcp"
	conn, err := net.Dial(protocol, addr)
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string

		for _, node := range n.knownPeers {
			if node.TcpAddress() != addr {
				updatedNodes = append(updatedNodes, node.TcpAddress())
			}
		}

		return err
	}
	defer conn.Close()

	if _, err = io.Copy(conn, bytes.NewReader(data)); err != nil {
		return err
	}

	return nil
}
