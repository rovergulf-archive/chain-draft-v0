package client

import (
	"context"
	"github.com/rovergulf/rbn/proto"
)

func (c *NetherClient) TxAdd(ctx context.Context, data []byte) ([]byte, error) {
	res, err := c.rpcClient.RpcCall(ctx, &proto.CallRequest{
		Cmd:  proto.CallRequest_TX_ADD,
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	c.logger.Debugf("Res status: %d (0 = Success!), Data len: %d", res.Status, len(res.Data))

	return res.Data, nil
}
