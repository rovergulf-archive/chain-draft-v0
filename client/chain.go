package client

import (
	"context"
	"github.com/rovergulf/rbn/proto"
)

func (c *NetherClient) SyncGenesis(ctx context.Context) error {
	res, err := c.rpcClient.RpcCall(ctx, &proto.CallRequest{
		Cmd:  proto.CallRequest_SYNC_GEN,
		Data: nil,
	})
	if err != nil {
		return err
	}

	c.logger.Debugf("Res status: %d (0 = Success!), Data len: %d", res.Status, len(res.Data))

	return nil
}
