package rosterssh

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// RosterItem holds the data relevant for an SSH configuration of a entry in the
// Salt roster file.
type RosterItem struct {
	Target     string
	Host       string
	User       string
	Port       string
	SSHOptions []string
}

// Set matches name to a specific RosterItem field and set value to it.
// If it doesn't match it returns false.
// the matching is case insensitive.
func (i *RosterItem) Set(name, value string) bool {
	switch strings.ToLower(name) {
	case "target":
		i.Target = value
	case "host":
		i.Host = value
	case "user":
		i.User = value
	case "port":
		i.Port = value
	case "sshoptions":
		i.SSHOptions = append(i.SSHOptions, value)
	default:
		return false
	}

	return true
}

// RosterParser parses a Salt roster file to extract from each target only the fields that are
// related to SSH client configuration.
//
// Use the constructor function to create one, the zero value isn't valid.
//
// See: https://docs.saltproject.io/en/latest/topics/ssh/roster.html
type RosterParser struct {
	scanner       *bufio.Scanner
	commentRegExp *regexp.Regexp
	err           error
	line          uint
	current       *RosterItem
	next          *RosterItem
}

// NewRosterParser creates a RosterParser for parsing the roster file content read through r.
//
// The commentPrefix is the prefix used in comments to set dynamic values; it's recommended to only
// use lowercase letters and numbers and be short.
// Comments that contains dynamic values are of the format `#{prefix} {key}: {value}`; they can be
// indented with any number of spaces or tabs.`{key}` is the field name that it has a dynamic value
// and `{value}` the value for this dynamic field that it should be replaced by the value specified
// by the user; for example `#ifc user: {MY_GCP_USER}` indicates that the SSH "user" is the
// `{MY_GCP_USER}` placeholder and the user has to provide the value for the placeholder. The space
// after the colon is optional, hence `#ifc user: {MY_GCP_USER}` and `#ifc user:{MY_GCP_USER}` are
// both valid and equivalent.
// If you don't use comments with prefixes, set a non-empty value.
//
// It returns an error if commentPrefix is empty or contains characters that avoid to compile a
// regular expression.
func NewRosterParser(r io.Reader, commentPrefix string) (*RosterParser, error) {
	if commentPrefix == "" {
		return nil, errors.New("the comment's prefix cannot be empty")
	}

	commentRegExp, err := regexp.Compile(`^[\s\t]*#` + commentPrefix + `\s(\w+):\s?(.+)`)
	if err != nil {
		return nil, fmt.Errorf(
			"the comment's prefix contains unappropriated characters for building a regular expression. %w",
			err,
		)
	}

	return &RosterParser{
		scanner:       bufio.NewScanner(r),
		commentRegExp: commentRegExp,
	}, nil
}

// Err returns an error if there was one or nil after Next returns false.
func (rp *RosterParser) Err() error {
	if rp.err != nil {
		return rp.err
	}
	return rp.scanner.Err()
}

// Item returns the current item. It returns a reference, so the caller should
// only use it before the next call to Next.
func (rp *RosterParser) Item() *RosterItem {
	return rp.current
}

// Next moves the cursor to the next item to be retrieved by the Item method.
func (rp *RosterParser) Next() bool {
	if rp.err != nil {
		return false
	}

	item := &RosterItem{}

	// rp.next isn't nil when the previous call to this method ended when reading a line that was the
	// beginning of the following target.
	if rp.next == nil {
		more := rp.scanner.Scan()
		for more {
			rp.line++

			l := rp.scanner.Text()
			if isBlankLineRegExp.MatchString(l) || isCommentRegExp.MatchString(l) {
				more = rp.scanner.Scan()
				continue
			}

			if isIndentedRegExp.MatchString(l) {
				rp.err = fmt.Errorf("invalid roster at line %d, expected a target ID, found an indented line", rp.line)
				return false
			}

			item.Target = strings.Trim(l, ": ")
			break
		}

		if !more {
			return false
		}
	} else {
		item = rp.next
		rp.next = nil
	}

	rp.current = item

	inSSHOptions := false

	// Loop until it finds a line which isn't blank, nor a comment, nor indented because then it's the
	// beginning of the following target.
	for rp.scanner.Scan() {
		rp.line++

		l := rp.scanner.Text()
		if isBlankLineRegExp.MatchString(l) {
			continue
		}

		if k, v, ok := rp.parseComment(l); ok {
			item.Set(k, v)
			continue
		}

		if isCommentRegExp.MatchString(l) {
			continue
		}

		if !isIndentedRegExp.MatchString(l) {
			rp.next = &RosterItem{Target: strings.Trim(l, ": ")}
			return true
		}

		// If the previous iteration was a block sequence for the ssh_options then
		// we expect to be an item
		if inSSHOptions {
			if v, ok := extractBlockSeqElem(l); ok {
				item.Set("sshoptions", strings.Trim(v, `" `))
				continue
			} else {
				inSSHOptions = false
			}
		}

		k, v, ok := extractTargetField(l)
		if !ok {
			continue
		}

		if k == "sshoptions" {
			// ssh_options field contains flow sequence, parse it.
			if v != "" {
				seq := parseFlowSeq(v)
				for _, s := range seq {
					item.Set(k, s)
				}
			} else {
				inSSHOptions = true
			}
		} else {
			item.Set(k, v)
		}
	}

	return true
}

func (rp *RosterParser) parseComment(s string) (key, val string, ok bool) {
	kv := rp.commentRegExp.FindStringSubmatch(s)
	if len(kv) != 3 {
		return "", "", false
	}

	return kv[1], strings.Trim(kv[2], "\t "), true
}

var (
	isCommentRegExp   = regexp.MustCompile(`^[\s\t]*#`)
	isIndentedRegExp  = regexp.MustCompile(`^[\s\r]+`)
	isBlankLineRegExp = regexp.MustCompile(`^[\s\t]*$`)
)

var extractTargetFieldRegExp = regexp.MustCompile(`^[\s\t]+(\w+):[\s\t]*(.*)`)

// extractTargetField extract form s the field's key, and value, and true, if s doesn't contain a
// field or the field isn't any of those related to SSH client configuration then it returns  false.
// If key has underscores they are removed.
//
// A field's value can be empty if the value is a string and empty or if it's a block sequence. In
// the case of flow sequence the value is the full list.
func extractTargetField(s string) (key, val string, ok bool) {
	kv := extractTargetFieldRegExp.FindStringSubmatch(s)
	if len(kv) != 3 {
		return "", "", false
	}

	k := strings.ReplaceAll(kv[1], "_", "")
	switch k {
	case "host":
	case "user":
	case "port":
	case "sshoptions":
	default:
		return "", "", false
	}

	return k, strings.Trim(kv[2], "\t "), true
}

var extractBlockSeqElemRegExp = regexp.MustCompile(`^[\s\t]+-[\s\t]*(.+)`)

func extractBlockSeqElem(s string) (val string, ok bool) {
	v := extractBlockSeqElemRegExp.FindStringSubmatch(s)
	if len(v) != 2 {
		return "", false
	}

	return v[1], true
}

func parseFlowSeq(s string) []string {
	l := strings.Trim(s, "[]")
	items := strings.Split(l, ",")

	for i := range items {
		items[i] = strings.Trim(items[i], `" `)
	}

	return items
}
