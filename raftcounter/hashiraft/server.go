package hashiraft

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

const (
	// NOTE: copied from consul
	snapshotsRetained = 2
	raftLogCacheSize  = 512
)

// Config defines the configurations for the counter node.
type Config struct {
	// Bootstrap defines if the server needed to bootstrap on startup.
	Bootstrap bool

	// RaftBindAddr defines the bind address for raft protocol.
	RaftBindAddr string

	// RPCBindAddr defines the bind address for rpc service.
	RPCBindAddr string

	// DataDir defines the data dir for the server.
	DataDir string

	// JoinAddress defines the join address for the server node.
	JoinAddress string
}

// DefaultConfig creates default config.
func DefaultConfig() *Config {
	return &Config{
		Bootstrap:    false,
		RaftBindAddr: "127.0.0.1:3333",
		RPCBindAddr:  "127.0.0.1:13333",
	}
}

// CreateServer creates a raft counter node from config.
func (c *Config) CreateServer() (*Server, error) {
	node := &Server{
		config:    c,
		rpcServer: rpc.NewServer(),
	}
	if node.config == nil {
		node.config = DefaultConfig()
	}

	if err := node.setupRaft(); err != nil {
		return nil, err
	}

	if err := node.setupRPC(); err != nil {
		return nil, err
	}

	return node, nil
}

// Server implements a raft server node.
type Server struct {
	config *Config

	fsm               *fsmServer
	raft              *raft.Raft
	raftConfig        *raft.Config
	raftAdvertiseAddr *net.TCPAddr

	rpcServer   *rpc.Server
	rpcListener net.Listener
}

func (s *Server) setupRaft() error {
	var err error

	s.raftConfig = raft.DefaultConfig()
	s.raftConfig.ProtocolVersion = raft.ProtocolVersionMax
	s.raftConfig.ShutdownOnRemove = true

	// setup transport layer
	bindAddr := s.config.RaftBindAddr
	advertiseAddr, err := net.ResolveTCPAddr("tcp", bindAddr)
	if err != nil {
		return err
	}
	transport, err := raft.NewTCPTransport(
		bindAddr,
		advertiseAddr,
		// TODO: tweak these configs
		3, time.Duration(60)*time.Second,
		// TODO; logger
		os.Stderr,
	)
	if err != nil {
		return err
	}
	s.raftAdvertiseAddr = advertiseAddr

	s.raftConfig.LocalID = raft.ServerID(s.config.RPCBindAddr)

	// setup fsm
	s.fsm = newFSM()

	// setup stores
	var (
		log    raft.LogStore
		stable raft.StableStore
		snap   raft.SnapshotStore
	)
	if s.config.DataDir == "" {
		// dev mode
		store := raft.NewInmemStore()
		stable = store
		log = store
		snap = raft.NewInmemSnapshotStore()
	} else {
		dir, err := prepareDataDir(s.config.DataDir)
		if err != nil {
			return err
		}

		store, err := raftboltdb.NewBoltStore(filepath.Join(dir, "raft.db"))
		if err != nil {
			return err
		}
		stable = store

		log, err = raft.NewLogCache(raftLogCacheSize, store)
		if err != nil {
			return err
		}

		// TODO: logger
		snap, err = raft.NewFileSnapshotStore(dir, snapshotsRetained, os.Stderr)
		if err != nil {
			return err
		}
	}

	if s.config.Bootstrap {
		// no need to join, try do bootstrap as leader
		hasState, err := raft.HasExistingState(log, stable, snap)
		if err != nil {
			return err
		}
		if !hasState {
			bootstrapConfig := raft.Configuration{
				Servers: []raft.Server{
					raft.Server{
						ID:      s.raftConfig.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			}
			err = raft.BootstrapCluster(s.raftConfig, log, stable, snap, transport, bootstrapConfig)
			if err != nil {
				return err
			}
		}
	}

	s.raft, err = raft.NewRaft(
		s.raftConfig,
		s.fsm,
		log,
		stable,
		snap,
		transport,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) setupRPC() error {
	s.rpcServer.Register(&Cluster{s})
	s.rpcServer.Register(&Counter{s})

	addr, err := net.ResolveTCPAddr("tcp", s.config.RPCBindAddr)
	if err != nil {
		return err
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	s.rpcListener = ln
	go s.rpcServer.Accept(s.rpcListener)
	// TODO: logger
	fmt.Printf("rpc server listening at: %s\n", addr)

	return nil
}

func (s *Server) Leader() string {
	return string(s.raft.Leader())
}

func (s *Server) isLeader() bool {
	return string(s.raft.Leader()) == s.config.RaftBindAddr
}

func decodeCounterApplyResponse(fut raft.ApplyFuture) (int64, error) {
	err := fut.Error()
	if err != nil {
		return 0, err
	}
	value, ok := fut.Response().(int64)
	if !ok {
		return 0, fmt.Errorf("invalid response value: %v", fut.Response())
	}
	return value, nil
}

func (s *Server) Incr() (int64, error) {
	if !s.isLeader() {
		rv := &CounterIncrResp{}
		err := s.rpc("Counter.Incr", &CounterIncrArgs{}, rv)
		return rv.CurrentValue, err
	}

	// TODO: tweak value
	timeout := time.Duration(10) * time.Second
	fut := s.raft.Apply(IncrCounterCommand.Encode(), timeout)

	return decodeCounterApplyResponse(fut)
}

func (s *Server) Decr() (int64, error) {
	if !s.isLeader() {
		rv := &CounterDecrResp{}
		err := s.rpc("Counter.Decr", &CounterDecrArgs{}, rv)
		return rv.CurrentValue, err
	}

	// TODO: tweak value
	timeout := time.Duration(10) * time.Second
	fut := s.raft.Apply(DecrCounterCommand.Encode(), timeout)

	return decodeCounterApplyResponse(fut)
}

func (s *Server) Current() (int64, error) {
	// TODO: check consistency
	return s.fsm.Counter().Current(), nil
}

func (s *Server) rpc(method string, args interface{}, reply interface{}) error {
	if s.raft == nil {
		return errors.New("not ready")
	}

	rpcAddr := s.config.JoinAddress

	configFut := s.raft.GetConfiguration()
	if err := configFut.Error(); err != nil {
		return err
	}
	leaderServerAddr := s.raft.Leader()
	if leaderServerAddr != "" {
		for _, server := range configFut.Configuration().Servers {
			if server.Address == leaderServerAddr {
				// NOTE: use server id as rpc address
				rpcAddr = string(server.ID)
				break
			}
		}
	}

	client, err := rpc.Dial("tcp", rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(method, args, reply)
}

func (s *Server) TryJoin() error {
	if s.config.JoinAddress == "" {
		// nothing to do
		return nil
	}

	args := &ClusterJoinArgs{
		ServerID:   s.config.RPCBindAddr,
		ServerAddr: s.config.RaftBindAddr,
	}
	reply := struct{}{}
	if err := s.rpc("Cluster.Join", args, &reply); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop() error {
	if s.raft == nil {
		return nil
	}

	if s.isLeader() {
		fut := s.raft.RemoveServer(s.raftConfig.LocalID, 0, 0).(raft.Future)
		if err := fut.Error(); err != nil {
			return nil
		}
	} else {
		reply := struct{}{}
		if err := s.rpc("Cluster.Leave", s.config.RPCBindAddr, &reply); err != nil {
			// ignore for now
			// TODO: logger
			fmt.Printf("leave cluster failed: %v\n", err)
		}
	}

	fut := s.raft.Shutdown()
	return fut.Error()
}
