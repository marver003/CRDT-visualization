package grpc

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/marver003/crdt/api/grpc/replication"
	"github.com/marver003/crdt/internal/crdt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func GossipServer(url string, counter *crdt.GCounter) {
	grpcServer := grpc.NewServer()

	gossipServer := NewGossipServer(counter)

	replication.RegisterReplicationServiceServer(grpcServer, gossipServer)

	hostName, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}

	fmt.Printf("gRPC server listening at %v%v\n", hostName, url)

	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}

type serverGossip struct {
	replication.UnimplementedReplicationServiceServer
	counter *crdt.GCounter
}

func NewGossipServer(counter *crdt.GCounter) *serverGossip {
	return &serverGossip{
		replication.UnimplementedReplicationServiceServer{},
		counter,
	}
}

func (s serverGossip) PushState(ctx context.Context, in *replication.PushStateRequest) (*emptypb.Empty, error) {
	// GCounter
	snapshot, err := s.counter.Unmarshal(in.State)

	if err == nil {
		s.counter.Merge(&snapshot)
	}

	return &emptypb.Empty{}, err
}
