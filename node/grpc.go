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
	case proto.Command_Sync:
		switch req.Entity {
		case proto.Entity_Genesis:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_Balance:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_Transaction:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_Block:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_BlockHeader:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_State:
			return nil, fmt.Errorf("not implemented")
		case proto.Entity_Peer:
			return n.handleRpcAddPeer(ctx, req.Data)
		default:
			return nil, fmt.Errorf("invalid entity")
		}
	case proto.Command_Add:
		switch req.Entity {
		case proto.Entity_Block:
			return n.handleRpcAddTx(ctx, req.Data)
		case proto.Entity_Transaction:
			return n.handleRpcAddTx(ctx, req.Data)
		default:
			return nil, fmt.Errorf("invalid entity")
		}
	case proto.Command_Get:
		return nil, fmt.Errorf("not implemented")
	case proto.Command_List:
		return nil, fmt.Errorf("not implemented")
	case proto.Command_Verify:
		return nil, fmt.Errorf("not implemented")
	case proto.Command_Drop:
		return nil, fmt.Errorf("not implemented")
	default:
		return nil, fmt.Errorf("invalid command")
	}
}
