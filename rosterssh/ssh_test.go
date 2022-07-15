package rosterssh

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSSHConfigEntry(t *testing.T) {
	t.Run("no users placeholders", func(t *testing.T) {
		const expEntry = `Host fleet-test-server
  Hostname 127.0.0.1
  Port 2222
  User test-1
  StrictHostChecking accept-new
  ProxyJump ubuntu@81.201.108.110:2222
  ControlPath ~/.ssh/mux/cm-%r@%h:%p
`
		item := RosterItem{
			Target: "server",
			Host:   "127.0.0.1",
			User:   "test-1",
			Port:   "2222",
			SSHOptions: []string{
				"ProxyJump=ubuntu@81.201.108.110:2222",
				"ControlPath=~/.ssh/mux/cm-%r@%h:%p",
			},
		}

		opts := SSHConfigOpts{
			ExtraSSHOptions: map[string]string{
				"StrictHostChecking": "accept-new",
			},
			Prefix: "fleet-test-",
		}

		buf := &bytes.Buffer{}
		require.NoError(t, writeSSHConfigEntry(item, opts, buf))
		require.Equal(t, expEntry, buf.String())
	})

	t.Run("no users placeholders and no port", func(t *testing.T) {
		const expEntry = `Host fleet-test-server
  Hostname 127.0.0.1
  User test-1
  StrictHostChecking accept-new
  ProxyJump ubuntu@81.201.108.110:2222
  ControlPath ~/.ssh/mux/cm-%r@%h:%p
`
		item := RosterItem{
			Target: "server",
			Host:   "127.0.0.1",
			User:   "test-1",
			SSHOptions: []string{
				"ProxyJump=ubuntu@81.201.108.110:2222",
				"ControlPath=~/.ssh/mux/cm-%r@%h:%p",
			},
		}

		opts := SSHConfigOpts{
			ExtraSSHOptions: map[string]string{
				"StrictHostChecking": "accept-new",
			},
			Prefix: "fleet-test-",
		}

		buf := &bytes.Buffer{}
		require.NoError(t, writeSSHConfigEntry(item, opts, buf))
		require.Equal(t, expEntry, buf.String())
	})

	t.Run("with user's placeholder", func(t *testing.T) {
		const expEntry = `Host fleet-test-server
  Hostname 127.0.0.1
  Port 2222
  User my-test-user
  StrictHostChecking accept-new
`
		item := RosterItem{
			Target: "server",
			Host:   "127.0.0.1",
			User:   "{YOUR_TEST_USER}",
			Port:   "2222",
		}

		opts := SSHConfigOpts{
			Prefix: "fleet-test-",
			ExtraSSHOptions: map[string]string{
				"StrictHostChecking": "accept-new",
			},
			UserPlaceholderValues: map[string]string{"{YOUR_TEST_USER}": "my-test-user"},
		}

		buf := &bytes.Buffer{}
		require.NoError(t, writeSSHConfigEntry(item, opts, buf))
		require.Equal(t, expEntry, buf.String())
	})
}
