package master

import (
	"log"
	"sync"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type Master struct {
	mu      sync.Mutex
	workers map[int]string
	nextId  int
}

func NewMaster(masterAddr string) *Master {
	return &Master{workers: make(map[int]string)}
}

func (m *Master) RegisterWorker(args *types.RegisterWorkerArgs, reply *types.RegisterWorkerReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextId++
	id := m.nextId
	m.workers[id] = args.Addr
	reply.WorkerID = id

	log.Printf("Woker %d registered in %s\n", id, args.Addr)
	return nil
}

func (m *Master) Heartbeat(args *types.HeartbeatArgs, reply *types.HeartbeatReply) error {
	log.Printf("Heartbeat received by worker %d\n", args.WorkerID)
	reply.Ack = true

	return nil
}
