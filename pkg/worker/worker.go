package worker

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

type MapFunc func(contents string) []types.KeyValue
type ReduceFunc func(values []string) string

func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

type Worker struct {
	id         int
	masterAddr string
	workerAddr string
	client     *rpc.Client
	mapF       MapFunc
	reduceF    ReduceFunc
}

func NewWorker(masterAddr, workerAddr string, mapF MapFunc, reduceF ReduceFunc) *Worker {
	if mapF == nil || reduceF == nil {
		log.Fatal("Map and Reduce functions cannot be nil")
	}
	return &Worker{
		masterAddr: masterAddr,
		workerAddr: workerAddr,
		mapF:       mapF,
		reduceF:    reduceF,
	}
}

func (w *Worker) Start() {
	client, err := rpc.Dial("tcp", w.masterAddr)
	if err != nil {
		log.Fatal("Error connecting to master:", err)
	}
	w.client = client

	args := &types.RegisterWorkerArgs{Addr: w.workerAddr}
	reply := &types.RegisterWorkerReply{}

	err = w.client.Call("Master.RegisterWorker", args, reply)
	if err != nil {
		log.Fatal("Error registering worker:", err)
	}

	w.id = reply.WorkerId
	log.Printf("Worker registered with id=%d", w.id)

	// commented for better visualization on terminal
	// go w.heartbeatLoop()

	go w.taskLoop()
}

func (w *Worker) RequestTask() *types.Task {
	args := &types.RequestTaskArgs{WorkerId: w.id}
	reply := &types.RequestTaskReply{}

	err := w.client.Call("Master.RequestTask", args, reply)
	if err != nil {
		log.Println("Error requesting task:", err)
		return nil
	}
	return reply.Task
}

func (w *Worker) ReportTask(task *types.Task) {
	args := &types.ReportTaskArgs{
		WorkerId: w.id,
		TaskId:   task.Id,
		TaskType: task.Type,
	}
	reply := &types.ReportTaskReply{}

	err := w.client.Call("Master.ReportTask", args, reply)
	if err != nil {
		log.Printf("Error reporting task %d: %v", task.Id, err)
	}

	if reply.Ack {
		log.Printf("Worker %d: reported task %d (%s) as completed", w.id, task.Id, task.Type)
	}
}

func (w *Worker) heartbeatLoop() {
	for {
		hbArgs := &types.HeartbeatArgs{WorkerId: w.id}
		hbReply := &types.HeartbeatReply{}

		err := w.client.Call("Master.Heartbeat", hbArgs, hbReply)
		if err != nil {
			log.Println("Heartbeat error, master might be down. Exiting.", err)
			os.Exit(1)
		}
		time.Sleep(2 * time.Second)
	}
}

func (w *Worker) doMapTask(task *types.Task) {
	content, err := os.ReadFile(task.Filename)
	if err != nil {
		log.Fatalf("Cannot read %v: %v", task.Filename, err)
	}

	kva := w.mapF(string(content))

	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Fatalf("Cannot create tmp directory: %v", err)
	}

	intermediateFiles := make([]*os.File, task.NReduce)
	encoders := make([]*json.Encoder, task.NReduce)
	for i := 0; i < task.NReduce; i++ {
		filename := fmt.Sprintf("%s/mr-intermediate-%d-%d", tmpDir, task.Id, i)
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("Cannot create intermediate file %s: %v", filename, err)
		}
		defer file.Close()
		intermediateFiles[i] = file
		encoders[i] = json.NewEncoder(file)
	}

	for _, kv := range kva {
		reduceTaskIndex := ihash(kv.Key) % task.NReduce
		err := encoders[reduceTaskIndex].Encode(&kv)
		if err != nil {
			log.Fatalf("Cannot write to intermediate file: %v", err)
		}
	}
}

func (w *Worker) doReduceTask(task *types.Task) {
	tmpDir := "tmp"
	intermediate := []types.KeyValue{}
	for i := 0; i < task.NMap; i++ {
		mapTaskIndex := i + 1
		filename := fmt.Sprintf("%s/mr-intermediate-%d-%d", tmpDir, mapTaskIndex, task.Id)
		file, err := os.Open(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			log.Fatalf("Could not open intermediate file %s: %v", filename, err)
		}
		dec := json.NewDecoder(file)
		for {
			var kv types.KeyValue
			if err := dec.Decode(&kv); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("Cannot decode intermediate file: %v", err)
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	sort.Slice(intermediate, func(i, j int) bool {
		return intermediate[i].Key < intermediate[j].Key
	})

	outputDir := "outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Cannot create output directory: %v", err)
	}
	oname := fmt.Sprintf("%s/mr-out-%d", outputDir, task.Id)
	ofile, _ := os.Create(oname)
	defer ofile.Close()

	i := 0
	for i < len(intermediate) {
		j := i + 1
		for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
			j++
		}
		values := []string{}
		for k := i; k < j; k++ {
			values = append(values, intermediate[k].Value)
		}
		output := w.reduceF(values)

		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)
		i = j
	}
}

func (w *Worker) taskLoop() {
	for {
		task := w.RequestTask()
		if task == nil {
			log.Printf("Worker %d: no task available, sleeping.", w.id)
			time.Sleep(2 * time.Second)
			continue
		}

		log.Printf("Worker %d: processing task %d (%s)", w.id, task.Id, task.Type)

		switch task.Type {
		case types.MapTaskType:
			w.doMapTask(task)
		case types.ReduceTaskType:
			w.doReduceTask(task)
		default:
			log.Printf("Worker %d: unknown task type %s", w.id, task.Type)
		}

		w.ReportTask(task)
	}
}
