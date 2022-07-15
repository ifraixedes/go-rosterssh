package rosterssh_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"storj.io/common/testcontext"

	"go.fraixed.es/rosterssh/rosterssh"
)

func TestRosterToSSHConfig(t *testing.T) {
	ctx := testcontext.New(t)
	f, err := os.Open("testdata/roster")
	require.NoError(t, err)
	defer ctx.Check(f.Close)

	rp, err := rosterssh.NewRosterParser(f, "ifc")
	require.NoError(t, err)

	ssh := &bytes.Buffer{}
	err = rosterssh.WriteSSHConfig(rp, rosterssh.SSHConfigOpts{
		Prefix: "test-",
		UserPlaceholderValues: map[string]string{
			"{YOUR_HETZNER_USERNAME}": "my-hetzner-username",
			"{YOUR_GCP_USERNAME}":     "my-google-cloud-user",
		},
	}, ssh)
	require.NoError(t, err)

	expSSH, err := ioutil.ReadFile("testdata/ssh-config")
	fmt.Printf("\n------\n%s\n-----\n", expSSH)
	require.NoError(t, err)
	require.Equal(t, string(expSSH), string(ssh.Bytes()))
}
