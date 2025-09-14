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
	masterAddr := flag.String("masterAddr", ":1234", "address for master RPC server")
	// nReduce := flag.Int("nReduce", 3, "number of reduce tasks")
	flag.Parse()

	master := &master.Master{Workers: make(map[int]string)}
	rpc.Register(master)

	l, err := net.Listen("tcp", *masterAddr)
	if err != nil {
		log.Fatal("listen error:", err)
	}

	fmt.Printf("Starting master at %s", *masterAddr)
	rpc.Accept(l)
}
