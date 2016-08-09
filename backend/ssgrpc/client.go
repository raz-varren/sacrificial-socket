package ssgrpc

import (
	"google.golang.org/grpc"
	"github.com/raz-varren/sacrificial-socket/backend/ssgrpc/transport"
	//"sync"
)

type propagateClient struct {
	conn   *grpc.ClientConn
	client transport.PropagateClient
}
