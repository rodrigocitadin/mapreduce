package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/rodrigocitadin/mapreduce/pkg/master"
	"github.com/rodrigocitadin/mapreduce/pkg/types"
	"github.com/rodrigocitadin/mapreduce/pkg/worker"
)

// mapF is the user-defined map function for the word count example.
// It splits a line into words and returns a KeyValue pair for each word.
func mapF(contents string) []types.KeyValue {
	ff := func(r rune) bool { return !unicode.IsLetter(r) }
	words := strings.FieldsFunc(contents, ff)

	kva := []types.KeyValue{}
	for _, w := range words {
		kva = append(kva, types.KeyValue{Key: w, Value: "1"})
	}
	return kva
}

// reduceF is the user-defined reduce function for the word count example.
// It sums the counts for a given word.
func reduceF(values []string) string {
	return fmt.Sprintf("%d", len(values))
}

// startMaster starts the master process.
func startMaster(masterAddr string) {
	// Clean up previous runs
	os.RemoveAll("tmp")
	os.RemoveAll("outputs")
	os.Mkdir("tmp", 0755)
	os.Mkdir("outputs", 0755)

	m := master.NewMaster()
	rpc.Register(m)

	l, err := net.Listen("tcp", masterAddr)
	if err != nil {
		log.Fatal("Master listen error:", err)
	}

	fmt.Printf("Master listening at %s\n", masterAddr)
	go rpc.Accept(l)

	// Wait for the job to complete before exiting.
	<-m.Done()
	log.Println("Master shutting down.")
}

// startWorker starts a worker process.
func startWorker(masterAddr, workerAddr string) {
	w := worker.NewWorker(masterAddr, workerAddr, mapF, reduceF)
	w.Start()
	// Keep the worker running
	select {}
}

// submitJob connects to the master and submits a new MapReduce job.
func submitJob(masterAddr, inputDir string) {
	client, err := rpc.Dial("tcp", masterAddr)
	if err != nil {
		log.Fatalf("Error connecting to master: %v", err)
	}

	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("Cannot read input directory: %v", err)
	}

	var inputFiles []string
	for _, file := range files {
		if !file.IsDir() {
			inputFiles = append(inputFiles, filepath.Join(inputDir, file.Name()))
		}
	}

	if len(inputFiles) == 0 {
		log.Fatalf("No input files found in %s", inputDir)
	}

	args := &types.SubmitJobArgs{
		InputFiles: inputFiles,
		NReduce:    2, // Example with 2 reduce tasks
	}
	reply := &types.SubmitJobReply{}

	err = client.Call("Master.SubmitJob", args, reply)
	if err != nil {
		log.Fatalf("Error submitting job: %v", err)
	}

	if reply.Ack {
		log.Printf("Job submitted successfully with ID %d", reply.JobId)
	} else {
		log.Println("Job submission failed")
	}
}

func main() {
	role := flag.String("role", "", "master, submit or worker")
	masterAddr := flag.String("masterAddr", ":9991", "master RPC address")
	workerAddr := flag.String("workerAddr", ":0", "worker RPC address (if role is worker)")
	inputDir := flag.String("inputs", "", "directory of input files (if role is submit)")
	flag.Parse()

	switch *role {
	case "master":
		startMaster(*masterAddr)
	case "submit":
		if *inputDir == "" {
			log.Fatal("Input directory must be specified for submit role")
		}
		submitJob(*masterAddr, *inputDir)
	case "worker":
		startWorker(*masterAddr, *workerAddr)
	default:
		fmt.Println("Usage: go run wordcount.go --role=<master|worker|submit> [options]")
		os.Exit(1)
	}
}
