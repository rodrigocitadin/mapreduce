package master

import (
	"errors"
	"log"
	"sync"

	"github.com/rodrigocitadin/mapreduce/pkg/scheduler"
	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type MasterPhase string

const (
	MapPhase    MasterPhase = "MAP"
	ReducePhase MasterPhase = "REDUCE"
	DonePhase   MasterPhase = "DONE"
)

type Master struct {
	mu                   sync.Mutex
	workers              map[int]string
	mapTasks             map[int]*types.Task
	reduceTasks          map[int]*types.Task
	taskQueue            *scheduler.TaskQueue
	nextWorkerId         int
	nMap                 int
	nReduce              int
	completedMapTasks    int
	completedReduceTasks int
	phase                MasterPhase
	done                 chan bool
}

func NewMaster() *Master {
	return &Master{
		workers:     make(map[int]string),
		mapTasks:    make(map[int]*types.Task),
		reduceTasks: make(map[int]*types.Task),
		taskQueue:   scheduler.NewTaskQueue(),
		phase:       MapPhase,
		done:        make(chan bool, 1),
	}
}

func (m *Master) RegisterWorker(args *types.RegisterWorkerArgs, reply *types.RegisterWorkerReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextWorkerId++
	id := m.nextWorkerId
	m.workers[id] = args.Addr
	reply.WorkerId = id

	log.Printf("Worker %d registered in %s\n", id, args.Addr)
	return nil
}

func (m *Master) Heartbeat(args *types.HeartbeatArgs, reply *types.HeartbeatReply) error {
	log.Printf("Heartbeat received by worker %d\n", args.WorkerId)
	reply.Ack = true

	return nil
}

func (m *Master) Done() <-chan bool {
	return m.done
}

func (m *Master) createReduceTasks() {
	// called with lock held
	for i := 0; i < m.nReduce; i++ {
		taskId := i // reduce task id is the partition number
		task := &types.Task{
			Id:      taskId,
			NReduce: m.nReduce,
			NMap:    m.nMap,
			Type:    types.ReduceTaskType,
			State:   types.IdleTaskState,
		}
		m.reduceTasks[taskId] = task
		m.taskQueue.Push(task)
	}
}

func (m *Master) CreateMapTasks(inputFiles []string, nReduce int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nMap = len(inputFiles)
	m.nReduce = nReduce

	for i, file := range inputFiles {
		taskId := i + 1
		task := &types.Task{
			Id:       taskId,
			NReduce:  nReduce,
			NMap:     m.nMap,
			Filename: file,
			Type:     types.MapTaskType,
			State:    types.IdleTaskState,
		}

		m.mapTasks[taskId] = task
		m.taskQueue.Push(task)
	}
}

func (m *Master) RequestTask(args *types.RequestTaskArgs, reply *types.RequestTaskReply) error {
	task := m.taskQueue.Pop()

	m.mu.Lock()
	defer m.mu.Unlock()

	task.State = types.InProgressTaskState
	task.WorkerId = args.WorkerId
	reply.Task = task

	return nil
}

func (m *Master) ReportTask(args *types.ReportTaskArgs, reply *types.ReportTaskReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var task *types.Task
	var ok bool

	if args.TaskType == types.MapTaskType {
		task, ok = m.mapTasks[args.TaskId]
	} else {
		task, ok = m.reduceTasks[args.TaskId]
	}

	if !ok {
		return errors.New("task not found")
	}

	if task.WorkerId != args.WorkerId {
		return errors.New("worker id does not match")
	}

	task.State = types.CompletedTaskState
	reply.Ack = true

	log.Printf("Task %d (%s) completed by worker %d\n", args.TaskId, args.TaskType, args.WorkerId)

	if args.TaskType == types.MapTaskType {
		m.completedMapTasks++
		if m.completedMapTasks == m.nMap {
			log.Println("All map tasks completed, starting reduce phase")
			m.phase = ReducePhase
			m.createReduceTasks()
		}
	} else {
		m.completedReduceTasks++
		if m.completedReduceTasks == m.nReduce {
			log.Println("All reduce tasks completed, job is done")
			m.phase = DonePhase
			m.done <- true
		}
	}

	return nil
}
