package raftcounter

// Counter defines counter interface.
type Counter interface {
	// Incr increases the counter, returns the increased value.
	Incr() (int64, error)
	// Decr decreases the counter, returns the decreased value.
	Decr() (int64, error)
	// Current retrieves current counter value. Default value is 0.
	Current() (int64, error)
}
