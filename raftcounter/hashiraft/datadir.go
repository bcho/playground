package hashiraft

import (
	"os"
	"path/filepath"
)

const (
	raftState = "raft"
)

func prepareDataDir(dir string) (string, error) {
	path := filepath.Join(dir, raftState)
	return path, os.MkdirAll(path, 0755)
}
