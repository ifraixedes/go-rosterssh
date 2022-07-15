package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"go.fraixed.es/rosterssh/rosterssh"
)

func main() {
	fl, err := parseFlags()
	if err != nil {
		exitWithError(err)
	}

	r, err := os.Open(fl.input)
	if err != nil {
		exitWithError(fmt.Errorf("impossible to open the Salt roster file. %w", err))
	}

	var w io.Writer
	if fl.output == "" {
		w = os.Stdout
	} else {
		w, err = os.Create(fl.output)
		if err != nil {
			exitWithError(fmt.Errorf("impossible to create output file. %w", err))
		}
	}

	rp, err := rosterssh.NewRosterParser(r, fl.prefixComment)
	if err != nil {
		exitWithError(err)
	}

	err = rosterssh.WriteSSHConfig(rp, rosterssh.SSHConfigOpts{
		Prefix:                fl.prefix,
		UserPlaceholderValues: fl.userValues,
		ExtraSSHOptions:       fl.extraOpts,
	}, w)
	if err != nil {
		exitWithError(err)
	}
}

type flags struct {
	prefix        string
	input         string
	output        string
	prefixComment string
	userValues    map[string]string
	extraOpts     map[string]string
}

func parseFlags() (flags, error) {
	prefix := pflag.String("prefix", "", "Prefix for each host found in the roster file")
	input := pflag.String("input", "", "File path to the Salt roster file")
	output := pflag.String("output", "", "File path to write the SSH configuration; when empty, it prints it out to stdout")
	prefixComment := pflag.String("prefix-comment", "",
		"Prefix used in the Roster comments to indicate user's specific values.\n"+
			"It cannot be empty even if the Roster doesn't contain any",
	)
	userValuesPairs := pflag.StringSlice("user-values", nil,
		"User's values to replace the placeholders set through the special comments\n"+
			"(the ones prefixed with prefix-comment). Values are name=value pairs",
	)

	extraOptsPairs := pflag.StringSlice("extra-opts", nil,
		"The extra SSH options to add to every SSH host generated from the Roster file.\n"+
			"This is useful if you want to use some SSH options that they are not present in the\n"+
			"Roster file and they won't be added there because they are specific to the user.\n"+
			"Values are name=value pairs",
	)

	pflag.Parse()

	if *input == "" {
		return flags{}, errors.New("input is required and cannot be empty")
	}

	var userValues map[string]string
	if len(*userValuesPairs) > 0 {
		userValues = make(map[string]string, len(*userValuesPairs))

		for _, pair := range *userValuesPairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				return flags{}, fmt.Errorf(
					"invalid user-values pair %q. It doesn't contain the `=` to split between field name and value",
					pair,
				)
			}

			userValues[kv[0]] = kv[1]
		}
	}

	var extraOpts map[string]string
	if len(*extraOptsPairs) > 0 {
		extraOpts = make(map[string]string, len(*extraOptsPairs))

		for _, pair := range *extraOptsPairs {
			kv := strings.SplitN(pair, "=", 2)
			if len(kv) != 2 {
				return flags{}, fmt.Errorf(
					"invalid extra options pair %q. It doesn't contain the `=` to split between field name and value",
					pair,
				)
			}

			extraOpts[kv[0]] = kv[1]
		}
	}

	return flags{
		prefix:        *prefix,
		input:         *input,
		output:        *output,
		prefixComment: *prefixComment,
		userValues:    userValues,
		extraOpts:     extraOpts,
	}, nil
}

func exitWithError(err error) {
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	os.Exit(1)
}
