package client

import (
	"testing"
	"time"

	"github.com/hashicorp/nomad/client/config"
	"github.com/hashicorp/nomad/helper/uuid"
	"github.com/hashicorp/nomad/nomad/structs"
	"github.com/hashicorp/nomad/testutil"
	"github.com/stretchr/testify/require"
)

func TestHearbeatStop_allocHook(t *testing.T) {
	t.Parallel()

	server, _, cleanupS1 := testServer(t, nil)
	defer cleanupS1()
	testutil.WaitForLeader(t, server.RPC)

	client, cleanupC1 := TestClient(t, func(c *config.Config) {
		c.RPCHandler = server
	})
	defer cleanupC1()

	d := 1 * time.Second
	alloc := &structs.Allocation{
		ID:        uuid.Generate(),
		TaskGroup: "foo",
		Job: &structs.Job{
			TaskGroups: []*structs.TaskGroup{
				{
					Name:                      "foo",
					StopAfterClientDisconnect: &d,
				},
			},
		},
		Resources: &structs.Resources{
			CPU:      100,
			MemoryMB: 100,
			DiskMB:   0,
		},
	}

	err := client.addAlloc(alloc, "")
	require.NoError(t, err)
	require.Contains(t, client.heartbeatStop.allocs, alloc.ID)

	err = client.registerNode()
	require.NoError(t, err)

	last, err := client.stateDB.GetLastHeartbeatOk()
	require.NoError(t, err)
	require.Empty(t, last)
}
