package main

import (
	"time"

	"github.com/bcho/playground/raftcounter/hashiraft"
)

func main() {
	config := hashiraft.DefaultConfig()
	_, err := config.CreateNode()
	if err != nil {
		panic(err)
	}

	for range time.Tick(time.Duration(3600) * time.Second) {
		// block
	}
}
