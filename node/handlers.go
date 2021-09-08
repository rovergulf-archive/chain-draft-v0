package node

import (
	"context"
	"encoding/json"
	"github.com/opentracing/opentracing-go"
	"github.com/rovergulf/rbn/pkg/traceutil"
)

func (n *Node) handleStatusMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	if n.tracer != nil {
		span := n.tracer.StartSpan("status_msg_handler", traceutil.ProvideParentSpan(ctx))
		defer span.Finish()
		ctx = opentracing.ContextWithSpan(ctx, span)
	}

	payload, err := json.Marshal(n.srv.NodeInfo())
	if err != nil {
		return nil, err
	}

	return &CallResult{
		Code: NodeDataMsg,
		Data: payload,
	}, nil
}

func (n *Node) handleNewBlockHashesMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleTransactionsMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleGetBlockHeadersMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleBlockHeadersMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) GetBlockBodiesMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) BlockBodiesMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleNewBlockMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleGetNodeDataMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleNodeDataMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleGetReceiptsMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleReceiptsMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleNewPooledTransactionHashesMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handleGetPooledTransactionsMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}

func (n *Node) handlePooledTransactionsMsg(ctx context.Context, payload []byte) (*CallResult, error) {
	return &CallResult{
		Code: 0,
		Data: nil,
	}, nil
}
