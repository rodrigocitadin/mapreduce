package main

import (
	"flag"
	"fmt"
)

func main() {
	masterAddr := flag.String("masterAddr", ":1234", "master RPC address")
	workerAddr := flag.String("workerAddr", ":0", "worker RPC address (optional)")
	flag.Parse()

	fmt.Println("Starting worker...")
	fmt.Println("Master:", *masterAddr)
	fmt.Println("Worker:", *workerAddr)
}
