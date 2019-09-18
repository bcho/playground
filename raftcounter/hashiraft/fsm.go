package hashiraft

import (
	"io"

	"github.com/hashicorp/raft"
)

type fsmServer struct {
}

func newFSM() *fsmServer {
	return &fsmServer{}
}

func (s *fsmServer) Apply(log *raft.Log) interface{} {
	// TODO
	return nil
}

func (s *fsmServer) Snapshot() (raft.FSMSnapshot, error) {
	// TODO
	return nil, nil
}

func (s *fsmServer) Restore(io.ReadCloser) error {
	// TODO
	return nil
}
