# Core

Core macros provided by [carapace-spec](https://github.com/rsteube/carapace-spec).

## directories

[`$directories`](https://rsteube.github.io/carapace/carapace/action/actionDirectories.html) completes directories.
```yaml
["$directories"]
```

## exec

[`$(<command>)`](https://rsteube.github.io/carapace/carapace/action/actionExecCommand.html) executes given command in a `sh` shell.

```yaml
["$(echo -e 'a\nb\nc')"]
```

## files

[`$files([<suffixes>])`](https://rsteube.github.io/carapace/carapace/action/actionFiles.html) completes files with an optional list of suffixes to filter on.

```yaml
["$files([.go, go.mod, go.sum])"]
```

## message

[`$message(<message>)`](https://rsteube.github.io/carapace/carapace/action/actionMessage.html) adds given error message to completion.

```yaml
["$message(some error)"]
```

## spec

`$spec(<file>)` completes arguments using the given spec file.
This implicitly [disables flag parsing](https://pkg.go.dev/github.com/spf13/cobra#Command) for the corresponding (sub)command.

```yaml
["$spec(example.yaml)"]
```
