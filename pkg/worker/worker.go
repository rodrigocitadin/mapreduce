package worker

import (
	"log"
	"net/rpc"
	"os"
	"time"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type Worker struct {
	id         int
	masterAddr string
	workerAddr string
	client     *rpc.Client
}

func NewWorker(masterAddr, workerAddr string) *Worker {
	return &Worker{
		masterAddr: masterAddr,
		workerAddr: workerAddr,
	}
}

func (w *Worker) Start() {
	client, err := rpc.Dial("tcp", w.masterAddr)
	if err != nil {
		log.Fatal("error connecting to master:", err)
	}
	w.client = client

	args := &types.RegisterWorkerArgs{Addr: w.workerAddr}
	reply := &types.RegisterWorkerReply{}

	err = w.client.Call("Master.RegisterWorker", args, reply)
	if err != nil {
		log.Fatal("error registering:", err)
	}

	w.id = reply.WorkerId
	log.Printf("Worker registered with id=%d", w.id)

	go w.heartbeatLoop()
	go w.taskLoop()
}

func (w *Worker) RequestTask() *types.Task {
	args := &types.RequestTaskArgs{WorkerId: w.id}
	reply := &types.RequestTaskReply{}

	err := w.client.Call("Master.RequestTask", args, reply)
	if err != nil {
		log.Fatal("Error requesting task:", err)
	}

	return reply.Task
}

func (w *Worker) ReportTask(task *types.Task) {
	args := &types.ReportTaskArgs{
		WorkerId: w.id,
		TaskId:   task.Id,
	}
	reply := &types.ReportTaskReply{}

	err := w.client.Call("Master.ReportTask", args, reply)
	if err != nil {
		log.Fatal("Error reporting task:", err)
	}

	if reply.Ack {
		log.Printf("Worker %d: reported task %d as completed", w.id, task.Id)
	}
}

func (w *Worker) heartbeatLoop() {
	for {
		hbArgs := &types.HeartbeatArgs{WorkerId: w.id}
		hbReply := &types.HeartbeatReply{}

		err := w.client.Call("Master.Heartbeat", hbArgs, hbReply)
		if err != nil {
			log.Println("heartbeat error:", err)
			os.Exit(1)
		}

		time.Sleep(2 * time.Second)
	}
}

func (w *Worker) taskLoop() {
	for {
		task := w.RequestTask()
		if task == nil {
			log.Printf("Worker %d: any task available, waiting...\n", w.id)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Worker %d: processing task %d (%s)", w.id, task.Id, task.Filename)
		time.Sleep(3 * time.Second) // simulating processing time
	}
}
