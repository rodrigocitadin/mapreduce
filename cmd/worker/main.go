package main

import (
	"flag"
	"fmt"

	"github.com/rodrigocitadin/mapreduce/pkg/worker"
)

func main() {
	masterAddr := flag.String("masterAddr", ":9991", "master RPC address")
	workerAddr := flag.String("workerAddr", ":0", "worker RPC address (optional)")
	flag.Parse()

	fmt.Printf("Starting worker at %s\n", *workerAddr)

	w := worker.NewWorker(*masterAddr, *workerAddr)
	w.Start()

	select {}
}
