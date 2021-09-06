package node

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/node/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// probably would be removed
func (n *Node) handleRpcRequest(ctx context.Context, data []byte) (*pb.CallResponse, error) {

	var req pb.CallRequest
	if err := proto.Unmarshal(data, &req); err != nil {
		n.logger.Errorf("Unable to unmarshal request: %s", err)
		return nil, err
	}
	n.logger.Debugw("Handling RPC request", "cmd", req.Cmd, "entity", req.Entity)

	switch req.Cmd {
	case pb.Command_Sync:
		switch req.Entity {
		case pb.Entity_Genesis:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_Balance:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_Transaction:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_Block:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_BlockHeader:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_State:
			return nil, fmt.Errorf("not implemented")
		case pb.Entity_Peer:
			return n.handleRpcAddPeer(ctx, req.Data)
		default:
			return nil, fmt.Errorf("invalid entity")
		}
	case pb.Command_Add:
		switch req.Entity {
		case pb.Entity_Block:
			return n.handleRpcAddTx(ctx, req.Data)
		case pb.Entity_Transaction:
			return n.handleRpcAddTx(ctx, req.Data)
		default:
			return nil, fmt.Errorf("invalid entity")
		}
	case pb.Command_Get:
		return nil, fmt.Errorf("not implemented")
	case pb.Command_List:
		return nil, fmt.Errorf("not implemented")
	case pb.Command_Verify:
		return nil, fmt.Errorf("not implemented")
	case pb.Command_Drop:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid command")
	}
}

func (n *Node) handleRpcAddPeer(ctx context.Context, data []byte) (*pb.CallResponse, error) {
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

	return &pb.CallResponse{
		Status: 0, // codes.OK
		Data:   nil,
	}, nil
}

func (n *Node) handleRpcGetBlock(ctx context.Context, data []byte) (*pb.CallResponse, error) {
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

	return &pb.CallResponse{
		Status: 0,
		Data:   result,
	}, nil
}

func (n *Node) handleRpcAddBlock(ctx context.Context, data []byte) (*pb.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	return &pb.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}

func (n *Node) handleRpcAddTx(ctx context.Context, data []byte) (*pb.CallResponse, error) {
	if len(data) < 0 {
		return nil, fmt.Errorf("empty data")
	}

	var req TxAddRequest
	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&req); err != nil {
		return nil, err
	}

	fmt.Println("got tx", req)

	return &pb.CallResponse{
		Status: 1,
		Data:   nil,
	}, fmt.Errorf("not implemented")
}
