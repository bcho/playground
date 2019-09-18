package hashiraft

import (
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
)

// Config defines the configurations for the counter node.
type Config struct {
	// RaftBindAddr defines the bind address for raft protocol.
	RaftBindAddr string
}

// DefaultConfig creates default config.
func DefaultConfig() *Config {
	return &Config{
		RaftBindAddr: "127.0.0.1:3333",
	}
}

// CreateNode creates a raft counter node from config.
func (c *Config) CreateNode() (*Node, error) {
	node := &Node{
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

type Node struct {
	config *Config

	fsm        *fsmServer
	raft       *raft.Raft
	raftConfig *raft.Config
}

func (n *Node) setupRaft() error {
	var err error

	n.raftConfig = raft.DefaultConfig()
	n.raftConfig.ProtocolVersion = raft.ProtocolVersionMax

	// setup transport layer
	bindAddr := n.config.RaftBindAddr
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

	n.raftConfig.LocalID = raft.ServerID(transport.LocalAddr())

	// setup fsm
	n.fsm = newFSM()

	// setup stores
	var (
		log    raft.LogStore
		stable raft.StableStore
		snap   raft.SnapshotStore
	)
	// TODO: use persistent store
	{
		store := raft.NewInmemStore()
		stable = store
		log = store
		snap = raft.NewInmemSnapshotStore()
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
						ID:      n.raftConfig.LocalID,
						Address: transport.LocalAddr(),
					},
				},
			}
			err = raft.BootstrapCluster(n.raftConfig, log, stable, snap, transport, bootstrapConfig)
			if err != nil {
				return err
			}
		}
	}

	n.raft, err = raft.NewRaft(
		n.raftConfig,
		n.fsm,
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
