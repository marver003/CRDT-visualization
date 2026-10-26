package simulator

import (
	"slices"

	"github.com/marver003/crdt/internal/crdt"
	"github.com/marver003/crdt/internal/types"
)

type PendingOpKey struct {
	To     types.ReplicaId
	Origin types.ReplicaId
	Seq    uint64
}

type PendingOp struct {
	Op       crdt.CounterOp
	ClientId types.ClientId
	ReqSeq   uint64
}

func (s *Simulator) trackOp(to types.ReplicaId, op crdt.CounterOp, clientId types.ClientId, reqSeq uint64) {
	s.PendingOps[PendingOpKey{To: to, Origin: op.Origin, Seq: op.Seq}] = &PendingOp{
		Op:       op,
		ClientId: clientId,
		ReqSeq:   reqSeq,
	}
}

func (s *Simulator) ackOp(to types.ReplicaId, op crdt.CounterOp) {
	delete(s.PendingOps, PendingOpKey{To: to, Origin: op.Origin, Seq: op.Seq})
}

func (s *Simulator) OnNodeEnabled(id types.ReplicaId) {
	if s.Store.CrdtType != types.GCounterOpType {
		return
	}

	peers := make([]types.ReplicaId, 0, len(s.Store.Nodes))

	for peer := range s.Store.Nodes {
		if peer != id && s.Store.IsUp(peer) {
			peers = append(peers, peer)
		}
	}

	slices.Sort(peers)

	for _, peer := range peers {
		s.Queue.Enqueue(SendSyncRequestOperation{
			BaseOperation: s.After(sendDelay),
			From:          id,
			To:            peer,
		})
	}

	keys := make([]PendingOpKey, 0, len(s.PendingOps))

	for key := range s.PendingOps {
		if key.Origin == id && s.Store.IsUp(key.To) {
			keys = append(keys, key)
		}
	}

	slices.SortFunc(keys, func(a, b PendingOpKey) int {
		if a.Seq != b.Seq {
			return int(a.Seq) - int(b.Seq)
		}
		return int(a.To[0]) - int(b.To[0])
	})

	for _, key := range keys {
		pending := s.PendingOps[key]
		op := pending.Op

		s.Queue.Enqueue(SendGossipMessageOperation{
			BaseOperation: s.After(sendDelay),
			From:          key.Origin,
			To:            key.To,
			ClientId:      pending.ClientId,
			Seq:           pending.ReqSeq,
			Op:            &op,
			Replay:        true,
		})
	}
}

func (op SendSyncRequestOperation) Execute(sim *Simulator) error {
	if !sim.Store.IsUp(op.From) {
		return nil
	}

	have, err := sim.Store.NodeHave(op.From)
	if err != nil {
		return err
	}

	deliverAt := sim.Now + sim.Timing.GossipLatency
	msg := sim.NewMessage(SyncRequestMessage, NodeAt(op.From), NodeAt(op.To), deliverAt)
	msg.Have = have

	sim.Queue.Enqueue(ReceiveSyncRequestOperation{
		BaseOperation: BaseOperation{StepId: deliverAt},
		Message:       msg,
	})

	return nil
}

func (op ReceiveSyncRequestOperation) Execute(sim *Simulator) error {
	sim.RemovePendingMessage(op.Message.Id)

	responder := types.ReplicaId(op.Message.To.Id)
	requester := types.ReplicaId(op.Message.From.Id)

	if !sim.Store.IsUp(responder) {
		return nil
	}

	ops, err := sim.Store.NodeOpsAfter(responder, op.Message.Have[responder])
	if err != nil {
		return err
	}

	for _, missing := range ops {
		missing := missing
		send := SendGossipMessageOperation{
			BaseOperation: sim.After(sendDelay),
			From:          responder,
			To:            requester,
			Op:            &missing,
			Replay:        true,
		}

		if pending, ok := sim.PendingOps[PendingOpKey{To: requester, Origin: missing.Origin, Seq: missing.Seq}]; ok {
			send.ClientId = pending.ClientId
			send.Seq = pending.ReqSeq
		}

		sim.Queue.Enqueue(send)
	}

	return nil
}
