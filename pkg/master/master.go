package master

import (
	"log"
	"sync"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type Master struct {
	Mu      sync.Mutex
	Workers map[int]string
	NextId  int
}

func (m *Master) RegisterWorker(args *types.RegisterWorkerArgs, reply *types.RegisterWorkerReply) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()

	m.NextId++
	id := m.NextId
	m.Workers[id] = args.Addr
	reply.WorkerID = id

	log.Printf("Woker %d registered in %s\n", id, args.Addr)
	return nil
}

func (m *Master) Heartbeat(args *types.HeartbeatArgs, reply *types.HeartbeatReply) error {
	log.Printf("Heartbeat received by worker %d\n", args.WorkerID)
	reply.Ack = true

	return nil
}
