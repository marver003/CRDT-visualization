package grpc

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/marver003/crdt/api/grpc/replication"
	"github.com/marver003/crdt/internal/crdt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type serverGossip struct {
	replication.UnimplementedReplicationServiceServer
	crdt crdt.CRDT
}

func GossipServer(url string, crdtInstance crdt.CRDT) {
	log.Println("Create Gossip server")
	grpcServer := grpc.NewServer()

	gossipServer := NewGossipServer(crdtInstance)

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

func NewGossipServer(crdtInstance crdt.CRDT) *serverGossip {
	return &serverGossip{
		replication.UnimplementedReplicationServiceServer{},
		crdtInstance,
	}
}

func (s serverGossip) PushState(ctx context.Context, in *replication.PushStateRequest) (*emptypb.Empty, error) {
	log.Printf("Received gossip request from CRDT ID: %v\n", in.CrdtId)

	if in.CrdtType != string(s.crdt.Type()) {
		return &emptypb.Empty{}, fmt.Errorf("crdt type mismatch: got %q, expected %q", in.CrdtType, s.crdt.Type())
	}

	snapshot, err := s.crdt.Unmarshal(in.State)

	if err == nil {
		s.crdt.Merge(snapshot)
	}

	return &emptypb.Empty{}, err
}
