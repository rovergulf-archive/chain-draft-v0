package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/rovergulf/rbn/proto"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"net"
)

func (n *Node) PrepareGrpcServer() (*grpc.Server, error) {

	var opts []grpc.ServerOption

	if n.tracer != nil {
		opts = append(opts, grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(n.tracer)))
	}

	if tlsConf, ok := viper.Get("tls.config").(tls.Config); ok {
		tlsConf.InsecureSkipVerify = false
		//credentials.NewServerTLSFromCert()
		bundle := credentials.NewServerTLSFromCert(&tlsConf.Certificates[0])
		opts = append(opts, grpc.Creds(bundle))
	}

	return grpc.NewServer(opts...), nil
}

func (n *Node) RunGrpcServer(addr string) error {
	var opts []grpc.ServerOption

	if n.tracer != nil {
		opts = append(opts, grpc.UnaryInterceptor(otgrpc.OpenTracingServerInterceptor(n.tracer)))
	}

	if viper.GetBool("grpc.tls.enabled") {
		certPath := viper.GetString("grpc.tls.cert")
		keyPath := viper.GetString("grpc.tls.key")
		bundle, err := credentials.NewServerTLSFromFile(certPath, keyPath)
		if err != nil {
			n.logger.Errorf("Unable to init server tls config: %s", err)
			return err
		} else {
			n.logger.Debugw("Loaded grpc tls config", "cert", certPath, "key", keyPath)
		}
		opts = append(opts, grpc.Creds(bundle))
	}

	n.grpcServer = grpc.NewServer(opts...)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		n.logger.Errorf("Unable to start net.Listener for [%s]: %s", "", err)
		return err
	}

	proto.RegisterNodeServiceServer(n.grpcServer, n)
	return n.grpcServer.Serve(l)
}

func (n *Node) Check(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	n.logger.Debug("Handle proto.Check request")
	var status proto.HealthCheckResponse_ServingStatus
	if n.grpcServer.GetServiceInfo() != nil {
		status = proto.HealthCheckResponse_SERVING
	} else {
		status = proto.HealthCheckResponse_NOT_SERVING
	}
	return &proto.HealthCheckResponse{
		Status: status,
	}, nil
}

func (n *Node) RpcCall(ctx context.Context, req *proto.CallRequest) (*proto.CallResponse, error) {
	n.logger.Debugw("RPC Call", "cmd", req.Cmd)

	switch req.Cmd {
	case proto.CallRequest_SYNC_PEERS:
		return n.handleRpcAddPeer(ctx, req.Data)
	case proto.CallRequest_SYNC_GEN:
		return nil, fmt.Errorf("not implemented")
	case proto.CallRequest_SYNC_STATE:
		return nil, fmt.Errorf("not implemented")
	case proto.CallRequest_SYNC_BLOCKS:
		return nil, fmt.Errorf("not implemented")
	//case rpc.CallRequest_BLOCK_ADD:
	//	return n.handleRpcAddBlock(ctx, req.Data)
	case proto.CallRequest_TX_ADD:
		return n.handleRpcAddTx(ctx, req.Data)
	case proto.CallRequest_TX_GET:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid command")
	}
}
