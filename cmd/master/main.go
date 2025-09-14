package main

import (
	"flag"
	"fmt"
)

func main() {
	masterAddr := flag.String("masterAddr", ":1234", "address for master RPC server")
	nReduce := flag.Int("nReduce", 3, "number of reduce tasks")
	flag.Parse()

	fmt.Println("Starting master...")
	fmt.Println("Addr:", *masterAddr)
	fmt.Println("Reducers:", *nReduce)
}
