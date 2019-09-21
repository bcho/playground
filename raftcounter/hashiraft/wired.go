package hashiraft

import (
	"fmt"

	"github.com/hashicorp/raft"
)

// TODO: extract package

type Counter struct {
	server *Server
}

type CounterIncrArgs struct{}

type CounterIncrResp struct {
	CurrentValue int64
}

func (c *Counter) Incr(args CounterIncrArgs, resp *CounterIncrResp) error {
	currentValue, err := c.server.Incr()
	if err != nil {
		return err
	}
	resp.CurrentValue = currentValue
	return nil
}

type CounterDecrArgs struct{}

type CounterDecrResp struct {
	CurrentValue int64
}

func (c *Counter) Decr(args CounterDecrArgs, resp *CounterDecrResp) error {
	currentValue, err := c.server.Decr()
	if err != nil {
		return err
	}
	resp.CurrentValue = currentValue
	return nil
}

type CounterCurrentArgs struct{}

type CounterCurrentResp struct {
	CurrentValue int64
}

func (c *Counter) Current(args CounterCurrentArgs, resp *CounterCurrentResp) error {
	currentValue, err := c.server.Current()
	if err != nil {
		return err
	}
	resp.CurrentValue = currentValue
	return nil
}

type Cluster struct {
	server *Server
}

type ClusterJoinArgs struct {
	ServerID   string
	ServerAddr string
}

type ClusterJoinResp struct{}

func (c *Cluster) Join(args *ClusterJoinArgs, rv *ClusterJoinResp) error {
	if !c.server.isLeader() && c.server.Leader() != "" {
		// forward to leader
		return c.server.rpc("Cluster.Join", args, rv)
	}

	configFut := c.server.raft.GetConfiguration()
	if err := configFut.Error(); err != nil {
		return err
	}

	srvID := raft.ServerID(args.ServerID)
	srvAddr := raft.ServerAddress(args.ServerAddr)
	for _, server := range configFut.Configuration().Servers {
		srvIDMatched := server.ID == srvID
		srvAddressMatched := server.Address == srvAddr
		switch {
		case !srvIDMatched && !srvAddressMatched:
			continue
		case srvIDMatched && srvAddressMatched:
			// already added
			return nil
		default:
			// remove the duplicated node
			removeFut := c.server.raft.RemoveServer(srvID, 0, 0)
			if err := removeFut.Error(); err != nil {
				return err
			}
		}
	}

	fut := c.server.raft.AddVoter(srvID, srvAddr, 0, 0).(raft.Future)
	return fut.Error()
}

type ClusterLeaveResp struct{}

func (c *Cluster) Leave(serverId string, rv *ClusterLeaveResp) error {
	if !c.server.isLeader() && c.server.Leader() != "" {
		// forward to leader
		return c.server.rpc("Cluster.Leave", serverId, rv)
	}

	srvID := raft.ServerID(serverId)
	fut := c.server.raft.RemoveServer(srvID, 0, 0)
	err := fut.Error()
	fmt.Printf("server %s left: %v\n", serverId, err)
	return err
}
