package rosterssh_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"storj.io/common/testcontext"

	"go.fraixed.es/rosterssh/rosterssh"
)

func TestRosterParser(t *testing.T) {
	expItems := []*rosterssh.RosterItem{
		{Target: "gcp-qa-thanos-0", Host: "10.226.119.18", User: "ivan"},
		{Target: "htz-client-ash-1", Host: "10.161.89.15", User: "root"},
		{Target: "htz-prod-cockroach-0", Host: "10.119.37.58", User: "{YOUR_HETZNER_USERNAME}"},
		{Target: "fr7-fw1", Host: "172.20.0.2", User: "{YOUR_GCP_USERNAME}", SSHOptions: []string{
			"ProxyJump=ubuntu@172.201.10.110:2222", "ControlPath=~/.ssh/mux/cm-%r@%h:%p",
		}},
		{Target: "fr7-fw2", Host: "172.20.0.3", User: "boom", SSHOptions: []string{
			"ProxyJump=ubuntu@10.201.108.111:7854",
		}},
		{Target: "fictional-box-1", Host: "10.22.119.18", SSHOptions: []string{
			"ProxyJump=ubuntu@172.201.10.11:333", "ControlPath=~/.ssh/mux/cm-%r@%h:%p",
		}},
		{Target: "fictional-box-2", Host: "172.226.19.188", User: "wow"},
	}

	ctx := testcontext.New(t)
	f, err := os.Open("testdata/roster")
	require.NoError(t, err)
	defer ctx.Check(f.Close)

	rp, err := rosterssh.NewRosterParser(f, "ifc")
	require.NoError(t, err)

	for i := 0; rp.Next(); i++ {
		item := rp.Item()
		assert.Equal(t, expItems[i], item)
	}

	require.NoError(t, rp.Err())
}
