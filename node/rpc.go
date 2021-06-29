package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/core"
	"github.com/rovergulf/rbn/rpc"
	"io"
	"log"
	"net"
)

type Addr struct {
	AddrList []string `json:"addr_list" yaml:"addr_list"`
}

type Block struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Block    []byte         `json:"block" yaml:"block"`
}

type GetBlocks struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
}

type GetData struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Type     string         `json:"type" yaml:"type"`
	ID       []byte         `json:"id" yaml:"id"`
}

type Inv struct {
	AddrFrom common.Address `json:"addr_from" yaml:"addr_from"`
	Type     string         `json:"type" yaml:"type"`
	Items    [][]byte       `json:"items" yaml:"items"`
}

type Tx struct {
	AddrFrom    common.Address `json:"addr_from" yaml:"addr_from"`
	Transaction []byte         `json:"transaction" yaml:"transaction"`
}

type TxAddReq struct {
	From    string `json:"from" yaml:"from"`
	FromPwd string `json:"from_pwd" yaml:"from_pwd"`
	To      string `json:"to" yaml:"to"`
	Value   uint64 `json:"value" yaml:"value"`
	Data    []byte `json:"data" yaml:"data"`
}

type Version struct {
	Version    int64          `json:"version" yaml:"version"`
	BestHeight int64          `json:"best_height" yaml:"best_height"`
	AddrFrom   common.Address `json:"addr_from" yaml:"addr_from"`
}

func (n *Node) SendTx(addr string, tnx *core.Transaction) error {
	serializedTx, err := tnx.Serialize()
	if err != nil {
		return err
	}

	data := Tx{
		AddrFrom:    common.HexToAddress(addr),
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

func (n *Node) handleRpcAddPeer(ctx context.Context, data []byte) (*rpc.CallResponse, error) {
	var pn PeerNode

	if err := pn.Deserialize(data); err != nil {
		return nil, err
	}
	fmt.Println("peer node arrived", pn)

	if n.tracer != nil {
		var opts []opentracing.StartSpanOption
		parentSpan := opentracing.SpanFromContext(ctx)
		if parentSpan != nil {
			opts = append(opts, opentracing.ChildOf(parentSpan.Context()))
		}
		span := n.tracer.StartSpan("add_peer", opts...)
		span.SetTag("addr", pn.TcpAddress())
		span.SetTag("account", pn.Account.Hex())
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	if err := n.addPeer(pn); err != nil {
		return nil, err
	}

	return &rpc.CallResponse{
		Status: 0, // codes.OK
		Data:   nil,
	}, nil
}

func (n *Node) handleRpcGetBlock(ctx context.Context, data []byte) (*rpc.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("invalid hash")
	}

	b, err := n.bc.GetBlock(common.BytesToHash(data))
	if err != nil {
		return nil, err
	}

	result, err := b.Serialize()
	if err != nil {
		return nil, err
	}

	return &rpc.CallResponse{
		Status: 0,
		Data:   result,
	}, nil
}

func (n *Node) handleRpcAddBlock(ctx context.Context, data []byte) (*rpc.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	return &rpc.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}

func (n *Node) handleRpcAddTx(ctx context.Context, data []byte) (*rpc.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	var req TxAddReq
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}

	fmt.Println("got tx", req)

	return &rpc.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}
