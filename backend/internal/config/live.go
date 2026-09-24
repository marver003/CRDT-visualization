package config

import (
	"errors"
	"fmt"
	"maps"
	"net"
	"slices"
	"strconv"

	"github.com/marver003/crdt/internal/types"
)

type Live struct {
	Mode         string                       `yaml:"mode"`
	NodeDefaults LiveNode                     `yaml:"node_defaults"`
	Nodes        map[types.ReplicaId]LiveNode `yaml:"nodes"`
}

type LiveNode struct {
	Crdt     types.CrdtType `yaml:"crdt"`
	Host     string         `yaml:"host"`
	HTTPPort int            `yaml:"http_port"`
	GRPCPort int            `yaml:"grpc_port"`
}

func LoadLive(path string) (*Live, error) {
	var cfg Live

	if err := load(path, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &cfg, nil
}

func (c *Live) Node(id types.ReplicaId) (LiveNode, bool) {
	node, exists := c.Nodes[id]
	if !exists {
		return LiveNode{}, false
	}

	if node.Crdt == "" {
		node.Crdt = c.NodeDefaults.Crdt
	}

	if node.Host == "" {
		node.Host = c.NodeDefaults.Host
	}

	return node, true
}

func (c *Live) Peers(id types.ReplicaId) []types.Peer {
	var peers []types.Peer

	for _, peerId := range slices.Sorted(maps.Keys(c.Nodes)) {
		if peerId == id {
			continue
		}

		peer, _ := c.Node(peerId)

		peers = append(peers, types.Peer{
			Id:       peerId,
			GRPCAddr: net.JoinHostPort(peer.Host, strconv.Itoa(peer.GRPCPort)),
		})
	}

	return peers
}

func (c *Live) validate() error {
	if c.Mode != "live" {
		return fmt.Errorf(`mode must be "live", got %q`, c.Mode)
	}

	if len(c.Nodes) == 0 {
		return errors.New("no nodes defined")
	}

	for id := range c.Nodes {
		node, _ := c.Node(id)

		if node.Crdt == "" {
			return fmt.Errorf("node %q: crdt is not set", id)
		}

		if node.Host == "" {
			return fmt.Errorf("node %q: host is not set", id)
		}

		if err := validatePort(fmt.Sprintf("node %q: http_port", id), node.HTTPPort); err != nil {
			return err
		}

		if err := validatePort(fmt.Sprintf("node %q: grpc_port", id), node.GRPCPort); err != nil {
			return err
		}
	}

	return nil
}
