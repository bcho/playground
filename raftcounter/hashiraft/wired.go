package hashiraft

// TODO: extract package

type Counter struct {
	server *Server
}

func (c *Counter) Incr() (int64, error) {
	return c.server.Incr()
}

func (c *Counter) Decr() (int64, error) {
	return c.server.Decr()
}

func (c *Counter) Current() (int64, error) {
	return c.server.Current()
}

type Cluster struct {
	server *Server
}

func (c *Cluster) Join(serverAddr string) error {
	// TODO
	return nil
}

func (c *Cluster) Leave(serverAddr string) error {
	// TODO
	return nil
}
