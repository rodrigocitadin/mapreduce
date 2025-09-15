package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"os"
	"path/filepath"

	"github.com/rodrigocitadin/mapreduce/pkg/master"
)

func main() {
	masterAddr := flag.String("masterAddr", ":9991", "address for master RPC server")
	flag.Parse()

	master := master.NewMaster()

	inputDir := "inputs"
	outputDir := "outputs"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("cannot create outputs directory: %v", err)
	}

	// Clean up previous output files for a fresh run
	outFiles, err := filepath.Glob(filepath.Join(outputDir, "mr-out-*"))
	if err == nil {
		for _, f := range outFiles {
			os.Remove(f)
		}
	}

	// Read input files from input directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		log.Fatalf("cannot read input directory: %v. Make sure it exists and contains input files.", err)
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

	master.CreateMapTasks(inputFiles, 2)

	rpc.Register(master)
	l, err := net.Listen("tcp", *masterAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("Starting master at %s\n", *masterAddr)

	go rpc.Accept(l)

	<-master.Done()
	fmt.Println("MapReduce job finished.")
}
