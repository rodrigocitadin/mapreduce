package master

import (
	"log"
	"sync"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type Master struct {
	mu       sync.Mutex
	workers  map[int]string
	mapTasks map[int]*types.Task
	nextId   int
}

func NewMaster(masterAddr string) *Master {
	return &Master{
		workers:  make(map[int]string),
		mapTasks: make(map[int]*types.Task),
	}
}

func (m *Master) RegisterWorker(args *types.RegisterWorkerArgs, reply *types.RegisterWorkerReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextId++
	id := m.nextId
	m.workers[id] = args.Addr
	reply.WorkerId = id

	log.Printf("Woker %d registered in %s\n", id, args.Addr)
	return nil
}

func (m *Master) Heartbeat(args *types.HeartbeatArgs, reply *types.HeartbeatReply) error {
	log.Printf("Heartbeat received by worker %d\n", args.WorkerId)
	reply.Ack = true

	return nil
}

func (m *Master) CreateMapTasks(inputFiles []string, nReduce int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	taskId := 0
	for _, file := range inputFiles {
		taskId++
		task := &types.Task{
			Id:       taskId,
			NReduce:  nReduce,
			Filename: file,
			Type:     types.MapTaskType,
			State:    types.IdleTaskState,
		}

		m.mapTasks[taskId] = task
	}
}

func (m *Master) RequestTask(args *types.RequestTaskArgs, reply *types.RequestTaskReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, task := range m.mapTasks {
		if task.State == types.IdleTaskState {
			task.State = types.InProgressTaskState
			task.WorkerId = args.WorkerId
			reply.Task = task

			return nil
		}
	}

	reply.Task = nil
	return nil
}
