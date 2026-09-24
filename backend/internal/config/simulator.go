package config

import (
	"errors"
	"fmt"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
	"go.yaml.in/yaml/v3"
)

const DefaultSimulatorPort = 8080

type Simulator struct {
	Mode      string            `yaml:"mode"`
	Simulator SimulatorSettings `yaml:"simulator"`
}

type SimulatorSettings struct {
	Port  int             `yaml:"port"`
	Crdt  types.CrdtType  `yaml:"crdt"`
	Nodes []SimulatorNode `yaml:"nodes"`
}

type SimulatorNode struct {
	Id    types.ReplicaId `yaml:"id"`
	State yaml.Node       `yaml:"state"`
}

func DefaultSimulator() *Simulator {
	return &Simulator{
		Mode: "simulator",
		Simulator: SimulatorSettings{
			Port: DefaultSimulatorPort,
			Crdt: types.GCounterType,
		},
	}
}

func LoadSimulator(path string) (*Simulator, error) {
	var cfg Simulator

	if err := load(path, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	return &cfg, nil
}

func (c *Simulator) Snapshots() (map[types.ReplicaId]crdt.Snapshot, error) {
	snapshots := make(map[types.ReplicaId]crdt.Snapshot, len(c.Simulator.Nodes))

	for _, node := range c.Simulator.Nodes {
		snapshot, err := decodeSnapshot(c.Simulator.Crdt, &node.State)
		if err != nil {
			return nil, fmt.Errorf("node %q: %w", node.Id, err)
		}

		snapshots[node.Id] = snapshot
	}

	return snapshots, nil
}

func decodeSnapshot(crdtType types.CrdtType, state *yaml.Node) (crdt.Snapshot, error) {
	switch crdtType {
	case types.GCounterType:
		var counts map[types.ReplicaId]uint64
		if err := decodeState(state, &counts); err != nil {
			return nil, err
		}

		return &crdt.GCounterSnapshot{Counts: counts}, nil

	case types.GSetType:
		var elements []string
		if err := decodeState(state, &elements); err != nil {
			return nil, err
		}

		set := make(types.Set[string], len(elements))
		for _, element := range elements {
			set[element] = struct{}{}
		}

		return &crdt.GSetSnapshot{Set: set}, nil

	default:
		return nil, fmt.Errorf("unsupported crdt type %q", crdtType)
	}
}

func decodeState(state *yaml.Node, out any) error {
	if state.IsZero() {
		return nil
	}

	if err := state.Decode(out); err != nil {
		return fmt.Errorf("invalid state: %w", err)
	}

	return nil
}

func (c *Simulator) validate() error {
	if c.Mode != "simulator" {
		return fmt.Errorf(`mode must be "simulator", got %q`, c.Mode)
	}

	if err := validatePort("simulator.port", c.Simulator.Port); err != nil {
		return err
	}

	if c.Simulator.Crdt == "" {
		return errors.New("simulator.crdt is not set")
	}

	seen := make(map[types.ReplicaId]bool, len(c.Simulator.Nodes))

	for _, node := range c.Simulator.Nodes {
		if node.Id == "" {
			return errors.New("simulator.nodes: every node needs an id")
		}

		if seen[node.Id] {
			return fmt.Errorf("simulator.nodes: duplicate node id %q", node.Id)
		}

		seen[node.Id] = true
	}

	if _, err := c.Snapshots(); err != nil {
		return err
	}

	return nil
}
