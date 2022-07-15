package rosterssh

import (
	"fmt"
	"io"
	"strings"
	"text/template"
)

// SSHConfigOpts are the options to use for SSH configuration generation.
type SSHConfigOpts struct {
	Prefix                string
	ExtraSSHOptions       map[string]string
	UserPlaceholderValues map[string]string
}

// WriteSSHConfig parse rp's items for generating an SSH configuration with opts and write it down
// to w.
func WriteSSHConfig(rp *RosterParser, opts SSHConfigOpts, w io.Writer) error {
	newLine := []byte("\n")
	for rp.Next() {
		item := rp.Item()
		if err := writeSSHConfigEntry(*item, opts, w); err != nil {
			return err
		}

		if _, err := w.Write(newLine); err != nil {
			return err
		}
	}

	return rp.Err()
}

type sshConfigEntry struct {
	Host     string
	Hostname string
	Port     string
	User     string

	// SSHOptions are the additional SSH keywords not represented by the above fields.
	// It's a slice for writing them in the order that they are specified. This will ease testing.
	SSHOptions [][2]string
}

var sshConfigEntryTpl = template.Must(template.New("").Parse(`Host {{.Host}}
  Hostname {{.Hostname}}
{{- if ne .Port "" }}
  Port {{.Port}}
{{- end }}
{{- if ne .User "" }}
  User {{.User}}
{{- end}}
{{- range $value :=  .SSHOptions }}
  {{index $value 0}} {{index $value 1}}
{{- end }}
`))

// writeSSHConfigEntry transforms item to an SSH section applying opts and writing it down into w.
// opts.ExtraSSHOptions are written before item.SSHOptions.
func writeSSHConfigEntry(item RosterItem, opts SSHConfigOpts, w io.Writer) error {
	entry := sshConfigEntry{
		Host:     opts.Prefix + item.Target,
		Hostname: item.Host,
		User:     item.User,
		Port:     item.Port,
	}

	if opts.UserPlaceholderValues != nil {
		if v, ok := opts.UserPlaceholderValues[item.User]; ok {
			entry.User = v
		}
	}

	var sshOptsKeys map[string]int
	if n := len(opts.ExtraSSHOptions) + len(item.SSHOptions); n > 0 {
		sshOptsKeys = make(map[string]int, n)
		entry.SSHOptions = make([][2]string, 0, n)
	}

	var i int
	for k, v := range opts.ExtraSSHOptions {
		sshOptsKeys[k] = i
		i++
		entry.SSHOptions = append(entry.SSHOptions, [2]string{k, v})
	}

	for _, o := range item.SSHOptions {
		kvo := strings.SplitN(o, "=", 2)
		if len(kvo) != 2 {
			return fmt.Errorf(
				"roster target %q contains an invalid SSH option because it doesn't have an '=': %q",
				item.Target, o,
			)
		}

		if i, ok := sshOptsKeys[kvo[0]]; ok {
			copy(entry.SSHOptions[i:], entry.SSHOptions[i+1:])
		}

		entry.SSHOptions = append(entry.SSHOptions, [2]string{kvo[0], kvo[1]})
	}

	return sshConfigEntryTpl.Execute(w, entry)
}
