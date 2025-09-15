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
	"strings"
	"time"
	"unicode"

	"github.com/rodrigocitadin/mapreduce/pkg/types"
)

// mapF is the user-defined map function.
// For word count, it returns a slice of KeyValue pairs, where each key is a word and the value is "1".
func mapF(contents string) []types.KeyValue {
	// function to determine if a character is a letter.
	ff := func(r rune) bool { return !unicode.IsLetter(r) }

	// split contents into a slice of words.
	words := strings.FieldsFunc(contents, ff)

	kva := []types.KeyValue{}
	for _, w := range words {
		kv := types.KeyValue{Key: w, Value: "1"}
		kva = append(kva, kv)
	}
	return kva
}

// reduceF is the user-defined reduce function.
// For word count, it sums the values for a given key.
func reduceF(values []string) string {
	// return the number of occurrences of this word.
	return fmt.Sprintf("%d", len(values))
}

// ihash uses the FNV-1a hash function to map a key to a reduce task.
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
		TaskType: task.Type,
	}
	reply := &types.ReportTaskReply{}

	err := w.client.Call("Master.ReportTask", args, reply)
	if err != nil {
		log.Fatal("Error reporting task:", err)
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
			log.Println("heartbeat error:", err)
			os.Exit(1)
		}

		time.Sleep(2 * time.Second)
	}
}

func (w *Worker) doMapTask(task *types.Task) {
	// Read input file content
	content, err := os.ReadFile(task.Filename)
	if err != nil {
		log.Fatalf("cannot read %v: %v", task.Filename, err)
	}

	// Apply map function
	kva := mapF(string(content))

	// Create intermediate files and encoders
	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		log.Fatalf("cannot create tmp directory: %v", err)
	}

	intermediateFiles := make([]*os.File, task.NReduce)
	encoders := make([]*json.Encoder, task.NReduce)
	for i := 0; i < task.NReduce; i++ {
		filename := fmt.Sprintf("%s/mr-intermediate-%d-%d", tmpDir, task.Id, i)
		file, err := os.Create(filename)
		if err != nil {
			log.Fatalf("cannot create intermediate file %s: %v", filename, err)
		}
		intermediateFiles[i] = file
		encoders[i] = json.NewEncoder(file)
	}

	// Write key-value pairs to intermediate files
	for _, kv := range kva {
		reduceTaskIndex := ihash(kv.Key) % task.NReduce
		err := encoders[reduceTaskIndex].Encode(&kv)
		if err != nil {
			log.Fatalf("cannot write to intermediate file: %v", err)
		}
	}

	// Close files
	for _, file := range intermediateFiles {
		file.Close()
	}
}

func (w *Worker) doReduceTask(task *types.Task) {
	// Read intermediate files
	tmpDir := "tmp"
	intermediate := []types.KeyValue{}
	for i := 0; i < task.NMap; i++ {
		mapTaskIndex := i + 1
		filename := fmt.Sprintf("%s/mr-intermediate-%d-%d", tmpDir, mapTaskIndex, task.Id)
		file, err := os.Open(filename)
		if err != nil {
			log.Printf("Could not open intermediate file %s: %v. It might not exist if no keys were mapped to it.", filename, err)
			continue
		}
		dec := json.NewDecoder(file)
		for {
			var kv types.KeyValue
			if err := dec.Decode(&kv); err == io.EOF {
				break
			} else if err != nil {
				log.Fatalf("cannot decode intermediate file: %v", err)
			}
			intermediate = append(intermediate, kv)
		}
		file.Close()
	}

	// Sort by key
	sort.Slice(intermediate, func(i, j int) bool {
		return intermediate[i].Key < intermediate[j].Key
	})

	// Create output file in the output directory
	outputDir := "outputs"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("cannot create output directory: %v", err)
	}
	oname := fmt.Sprintf("%s/mr-out-%d", outputDir, task.Id)
	ofile, _ := os.Create(oname)

	// Group by key and apply reduce function
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
		output := reduceF(values)

		// Write the result to the output file.
		fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

		i = j
	}

	ofile.Close()
}

func (w *Worker) taskLoop() {
	for {
		task := w.RequestTask()
		if task == nil {
			log.Printf("Worker %d: no task available, system might be done.", w.id)
			time.Sleep(2 * time.Second)
			// In a real scenario, the master might tell the worker to shut down.
			// For now, the worker will keep polling.
			continue
		}

		log.Printf("Worker %d: processing task %d (%s) file: %s", w.id, task.Id, task.Type, task.Filename)

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
