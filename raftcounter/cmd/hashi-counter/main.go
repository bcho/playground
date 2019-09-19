package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/bcho/playground/raftcounter/hashiraft"
)

var (
	flagRaftBindAddr = flag.String("raft-addr", "127.0.0.1:3333", "raft communication address")
)

func main() {
	flag.Parse()

	config := hashiraft.DefaultConfig()
	config.RaftBindAddr = *flagRaftBindAddr

	server, err := config.CreateServer()
	if err != nil {
		panic(err)
	}

	for range time.Tick(time.Duration(2) * time.Second) {
		updated, err := server.Incr()
		if err != nil {
			panic(fmt.Errorf("incr counter: %v", err))
		}
		current, err := server.Current()
		if err != nil {
			panic(fmt.Errorf("current counter: %v", err))
		}
		fmt.Printf("counter: updated=%d current=%d\n", updated, current)
	}
}
