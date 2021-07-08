package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (n *Node) handleRpcAddPeer(ctx context.Context, data []byte) (*proto.CallResponse, error) {
	var req JoinPeerRequest

	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}
	pn := req.From

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

	if _, ok := n.knownPeers.GetPeer(pn.TcpAddress()); ok {
		return nil, status.Errorf(codes.AlreadyExists, "Peer already known")
	}

	if err := n.addDbPeer(pn); err != nil {
		return nil, err
	}

	n.knownPeers.AddPeer(pn.TcpAddress(), pn)

	return &proto.CallResponse{
		Status: 0, // codes.OK
		Data:   nil,
	}, nil
}

func (n *Node) handleRpcGetBlock(ctx context.Context, data []byte) (*proto.CallResponse, error) {
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

	return &proto.CallResponse{
		Status: 0,
		Data:   result,
	}, nil
}

func (n *Node) handleRpcAddBlock(ctx context.Context, data []byte) (*proto.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	return &proto.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}

func (n *Node) handleRpcAddTx(ctx context.Context, data []byte) (*proto.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	var req TxAddRequest
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}

	fmt.Println("got tx", req)

	return &proto.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}
