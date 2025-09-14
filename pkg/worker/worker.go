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

	args := &types.RegisterWorkerArgs{Addr: w.workerAddr}
	reply := &types.RegisterWorkerReply{}

	err = client.Call("Master.RegisterWorker", args, reply)
	if err != nil {
		log.Fatal("error registering:", err)
	}

	w.id = reply.WorkerID
	log.Printf("Worker registered with id=%d", w.id)

	go func() {
		for {
			hbArgs := &types.HeartbeatArgs{WorkerID: w.id}
			hbReply := &types.HeartbeatReply{}

			err := client.Call("Master.Heartbeat", hbArgs, hbReply)
			if err != nil {
				log.Println("heartbeat error:", err)
				os.Exit(1)
			}

			time.Sleep(2 * time.Second)
		}
	}()
}
