package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/bcho/playground/raftcounter/hashiraft"
)

var (
	flagRaftBindAddr = flag.String("raft-addr", "127.0.0.1:3333", "raft communication address")
	flagRPCBindAddr  = flag.String("rpc-addr", "127.0.0.1:13333", "rpc service address")
	flagJoin         = flag.String("join", "", "address to join on startup")
	flagDataDir      = flag.String("data-dir", "", "data storage dir, defaults to use in-memory storage")
)

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.Parse()

	config := hashiraft.DefaultConfig()
	config.RaftBindAddr = *flagRaftBindAddr
	config.RPCBindAddr = *flagRPCBindAddr
	config.DataDir = *flagDataDir
	config.JoinAddress = *flagJoin
	config.Bootstrap = config.JoinAddress == ""

	server, err := config.CreateServer()
	if err != nil {
		panic(err)
	}
	if err := server.TryJoin(); err != nil {
		panic(err)
	}

	stopTickerChan := make(chan struct{})
	go func(stopChan chan struct{}) {
		for range time.Tick(time.Duration(2) * time.Second) {
			select {
			case <-stopChan:
				fmt.Println("ticker stopped")
				close(stopChan)
				return
			default:
			}

			leader := server.Leader()
			if leader == "" {
				fmt.Println("cluster is not ready...")
				continue
			}

			var (
				action  string
				updated int64
				err     error
			)
			switch rand.Int() % 2 {
			case 0:
				action = "incr"
				updated, err = server.Incr()
			case 1:
				action = "decr"
				updated, err = server.Decr()
			default:
			}
			if err != nil {
				fmt.Printf("update counter faild: %v\n", err)
				continue
			}
			current, err := server.Current()
			if err != nil {
				fmt.Printf("get current counter failed: %v\n", err)
				continue
			}

			fmt.Printf(
				"leader=%s current=%s %s counter: updated=%d current=%d\n",
				leader, config.RaftBindAddr,
				action, updated, current,
			)
		}
	}(stopTickerChan)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt)
	<-terminate

	stopTickerChan <- struct{}{}
	<-stopTickerChan

	if err := server.Stop(); err == nil {
		fmt.Println("server stopped")
	} else {
		panic(err)
	}
}
