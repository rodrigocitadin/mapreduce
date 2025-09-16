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
	IdlePhase   MasterPhase = "IDLE"
	MapPhase    MasterPhase = "MAP"
	ReducePhase MasterPhase = "REDUCE"
	DonePhase   MasterPhase = "DONE"
)

type Master struct {
	mu           sync.Mutex
	workers      map[int]string
	nextWorkerId int
	nextJobId    int
	done         chan bool

	// Job-specific state
	phase                MasterPhase
	mapTasks             map[int]*types.Task
	reduceTasks          map[int]*types.Task
	taskQueue            *scheduler.TaskQueue
	nMap                 int
	nReduce              int
	completedMapTasks    int
	completedReduceTasks int
}

func NewMaster() *Master {
	m := &Master{
		workers:     make(map[int]string),
		phase:       IdlePhase,
		taskQueue:   scheduler.NewTaskQueue(),
		done:        make(chan bool, 1), // Add a buffer to prevent deadlock
		mapTasks:    make(map[int]*types.Task),
		reduceTasks: make(map[int]*types.Task),
	}
	return m
}

func (m *Master) Done() <-chan bool {
	return m.done
}

func (m *Master) RegisterWorker(args *types.RegisterWorkerArgs, reply *types.RegisterWorkerReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextWorkerId++
	id := m.nextWorkerId
	m.workers[id] = args.Addr
	reply.WorkerId = id

	log.Printf("Worker %d registered from %s\n", id, args.Addr)
	return nil
}

func (m *Master) Heartbeat(args *types.HeartbeatArgs, reply *types.HeartbeatReply) error {
	log.Printf("Heartbeat received from worker %d\n", args.WorkerId)
	reply.Ack = true
	return nil
}

func (m *Master) SubmitJob(args *types.SubmitJobArgs, reply *types.SubmitJobReply) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.phase != IdlePhase {
		return errors.New("master is busy with another job")
	}

	log.Println("New job received. Creating Map tasks...")
	m.phase = MapPhase
	m.nMap = len(args.InputFiles)
	m.nReduce = args.NReduce

	for i, file := range args.InputFiles {
		taskId := i + 1
		task := &types.Task{
			Id:       taskId,
			NReduce:  m.nReduce,
			NMap:     m.nMap,
			Filename: file,
			Type:     types.MapTaskType,
			State:    types.IdleTaskState,
		}
		m.mapTasks[taskId] = task
		m.taskQueue.Push(task)
	}

	m.nextJobId++
	reply.JobId = m.nextJobId
	reply.Ack = true

	log.Printf("Job %d started with %d Map tasks and %d Reduce tasks.", reply.JobId, m.nMap, m.nReduce)

	return nil
}

func (m *Master) createReduceTasks() {
	for i := 0; i < m.nReduce; i++ {
		taskId := i
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

func (m *Master) RequestTask(args *types.RequestTaskArgs, reply *types.RequestTaskReply) error {
	task := m.taskQueue.Pop()

	m.mu.Lock()
	defer m.mu.Unlock()

	if task == nil {
		reply.Task = nil
		return nil
	}

	task.State = types.InProgressTaskState
	task.WorkerId = args.WorkerId
	reply.Task = task

	log.Printf("Assigned task %d (%s) to worker %d", task.Id, task.Type, task.WorkerId)

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

	if task.State == types.CompletedTaskState {
		log.Printf("Received duplicate completion report for task %d from worker %d", args.TaskId, args.WorkerId)
		reply.Ack = true
		return nil
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
			log.Println("All reduce tasks completed, job is done.")
			m.phase = DonePhase
			// Signal that the job is done, so the main process can exit.
			m.done <- true
		}
	}

	return nil
}
