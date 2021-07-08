package client

import (
	"context"
	"github.com/rovergulf/rbn/proto"
)

func (c *NetherClient) JoinKnownPeer(ctx context.Context, req []byte) ([]byte, error) {

	res, err := c.rpcClient.RpcCall(ctx, &proto.CallRequest{
		Cmd:  proto.CallRequest_SYNC_PEERS,
		Data: req,
	})
	if err != nil {
		return nil, err
	}

	return res.Data, nil
}
