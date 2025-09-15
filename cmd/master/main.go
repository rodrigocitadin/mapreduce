package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"

	"github.com/rodrigocitadin/mapreduce/pkg/master"
)

func main() {
	masterAddr := flag.String("masterAddr", ":9991", "address for master RPC server")
	flag.Parse()

	master := master.NewMaster(*masterAddr)

	inputFiles := []string{"input1.txt", "input2.txt", "input3.txt"}
	master.CreateMapTasks(inputFiles, 2)

	rpc.Register(master)
	l, err := net.Listen("tcp", *masterAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("Starting master at %s", *masterAddr)
	rpc.Accept(l)
}
