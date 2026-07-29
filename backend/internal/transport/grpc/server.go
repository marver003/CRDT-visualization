package grpc

import (
	"context"
	"log"
	"net"

	"github.com/marver003/crdt/api/grpc/replication"
	"github.com/marver003/crdt/internal/crdt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverGossip struct {
	replication.UnimplementedReplicationServiceServer
	counter *crdt.GCounter
}

func GossipServer(url string, counter *crdt.GCounter) {
	log.Println("Create Gossip server")
	grpcServer := grpc.NewServer()

	gossipServer := NewGossipServer(counter)

	replication.RegisterReplicationServiceServer(grpcServer, gossipServer)

	listener, err := net.Listen("tcp", url)
	if err != nil {
		panic(err)
	}

	log.Printf("gRPC server listening at localhost%v\n", url)

	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
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

	log.Printf("Received gossip request from CRDT ID: %v\n", in.CrdtId)

	if err == nil {
		s.counter.Merge(&snapshot)
	}

	return &emptypb.Empty{}, err
}
