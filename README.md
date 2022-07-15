# Roster SSH

Generate a SSH configuration file from the content of a [Salt Roster file](https://docs.saltproject.io/en/latest/topics/ssh/roster.html).

## Background

I had the need to have an SSH configuration file with all the hosts present in a [Salt Roster file](https://docs.saltproject.io/en/latest/topics/ssh/roster.html)
because I sometimes have to connect to a fleet of servers managed by [Salt over SSH](https://docs.saltproject.io/en/latest/topics/ssh/index.html).

But

- I don't want to all those servers manually.
- I want to have them in a separated SSH configuration file and included it to my default SSH
  configuration file through the `Include` keyword.
- I want all these hosts to have a prefix so it's easy to only see the list of them when I type
  `ssh prefix-` on my terminal and I press tab. But I also want this prefix to be configurable in
  case that I have more than one fleet of servers managed by Salt and belonging to different owners.
- I want to be able to set some comments in the Roster file for setting values that are specific to
  each users, for example, logging into one server may require the user's account to a cloud
  provider.
- I want to be able to add SSH options to each entry that they aren't set in the Roster file.

## Rationale

This package contain a command-line tool in the root of the repository that fulfills everything that
I want and I listed in the [Background section](#background).

It also expose all the functionality as a package that can be used by others under the `rosterssh`
directory.

## How to use it

To install the command-line tool `go install go.fraixed.es/rosterssh`.

To use the it as library then `go get go.fraixed.es/rosterssh` and import the package
`go.fraixed.es/rosterssh/rosterssh` (I know that it stutters but I didn't find a better name for the
package without having a name that it doesn't match the path with the package name and it doesn't
has a clueless name like `lib`).

The command-line tool has it's own `help` message, although, it mostly mentions the list of flags.
What it does, it's explained in this document.

Do you want to give it a run? You can execute the command-line tool with the Roster file inside of
the directory `rosterssh/testdata`.

The package it's documented and with the introduction in this document it should be easy to
understand how to use it besides you can have a look to the tests; aren't them enough for you?
Please, open an issue.

### How to set comments with user's specific values in a Roster file

In your roster you add a comment with a prefix that you'd like to use for them. This prefix can be
anything that it doesn't break a regular expression compilation. My recommendation is to be short
and only contain lowercase letters.

For example

```
htz-prod-cockroach-0:
  host: 10.119.37.58
  #ifc user: {YOUR_HETZNER_USERNAME}
  minion_opts:
    sops_pillars:
      - htz-prod-stargate.enc.yml
      - wireguard.enc.yml

```

In that example the _prefix_ is `ifc`.

This comments can be indented as you wish, mostly, any number of spaces and tabs before them are
ignored. The important part is that the prefix must be right after the `#` of the comment and it
must not be any space in the middle. After the prefix, you must use a space and then the SSH field
that you pretend to set with the with a specific user's values, although at this moment the only one
that makes sense is `user`.

After the field's name you must add a colon and after it the placeholder. A space between the colon
and placeholders is allowed but not mandatory, although, I would recommend to add it because it's
more readable.

When you use the command-line, you'll have to pass the value of the placeholder through a specific
flag for this purpose, for example `{YOUR_HETZNER_USERNAME}=ivan`. The placeholder can be anything,
it doesn't have to be surrounded by `{}`.
