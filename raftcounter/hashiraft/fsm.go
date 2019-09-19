package hashiraft

import (
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type fsmServer struct {
	counterLock *sync.RWMutex
	counter     *counter
}

func newFSM() *fsmServer {
	return &fsmServer{
		counterLock: &sync.RWMutex{},
		counter:     newCounter(),
	}
}

func (s *fsmServer) Counter() *counter {
	s.counterLock.RLock()
	defer s.counterLock.RUnlock()
	return s.counter
}

func (s *fsmServer) Apply(log *raft.Log) interface{} {
	buf := log.Data
	commandType := CommandType(buf[0])

	switch commandType {
	case IncrCounterCommand:
		return s.Counter().Incr()
	case DecrCounterCommand:
		return s.Counter().Decr()
	default:
		// unknown commmands
		panic(fmt.Errorf("failed to apply command: %v", buf))
	}
}

func (s *fsmServer) Snapshot() (raft.FSMSnapshot, error) {
	var err error

	counter := s.Counter()

	snapshot := &fsmSnapshot{}
	snapshot.counterValue, err = counter.EncodeValue()
	if err != nil {
		return nil, err
	}

	return snapshot, nil
}

func (s *fsmServer) Restore(rc io.ReadCloser) error {
	newCounter := newCounter()
	if err := newCounter.DecodeValue(rc); err != nil {
		return err
	}

	s.counterLock.Lock()
	s.counter = newCounter
	s.counterLock.Unlock()

	return nil
}

type fsmSnapshot struct {
	counterValue []byte
}

func (f *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(f.counterValue); err != nil {
		sink.Cancel()
		return err
	}

	return sink.Close()
}

func (f *fsmSnapshot) Release() {}
