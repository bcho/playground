package main

import (
	"flag"
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

	_, err := config.CreateNode()
	if err != nil {
		panic(err)
	}

	for range time.Tick(time.Duration(3600) * time.Second) {
		// block
	}
}
