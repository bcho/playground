package hashiraft

import (
	"fmt"
	"net"
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
	// RaftBindAddr defines the bind address for raft protocol.
	RaftBindAddr string

	// DataDir defines the data dir for the server.
	DataDir string
}

// DefaultConfig creates default config.
func DefaultConfig() *Config {
	return &Config{
		RaftBindAddr: "127.0.0.1:3333",
	}
}

// CreateServer creates a raft counter node from config.
func (c *Config) CreateServer() (*Server, error) {
	node := &Server{
		config: c,
	}
	if node.config == nil {
		node.config = DefaultConfig()
	}

	if err := node.setupRaft(); err != nil {
		return nil, err
	}

	return node, nil
}

// Server implements a raft server node.
type Server struct {
	config *Config

	fsm        *fsmServer
	raft       *raft.Raft
	raftConfig *raft.Config
}

func (s *Server) setupRaft() error {
	var err error

	s.raftConfig = raft.DefaultConfig()
	s.raftConfig.ProtocolVersion = raft.ProtocolVersionMax

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

	s.raftConfig.LocalID = raft.ServerID(transport.LocalAddr())

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

	// FIXME: join
	{
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

func (s *Server) Leader() string {
	return string(s.raft.Leader())
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
	// TODO: forward
	// TODO: tweak value
	timeout := time.Duration(10) * time.Second
	fut := s.raft.Apply(IncrCounterCommand.Encode(), timeout)

	return decodeCounterApplyResponse(fut)
}

func (s *Server) Decr() (int64, error) {
	// TODO: forward
	// TODO: tweak value
	timeout := time.Duration(10) * time.Second
	fut := s.raft.Apply(DecrCounterCommand.Encode(), timeout)

	return decodeCounterApplyResponse(fut)
}

func (s *Server) Current() (int64, error) {
	// TODO: check consistency
	return s.fsm.Counter().Current(), nil
}
