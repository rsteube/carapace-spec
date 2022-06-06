# Modifier

Modifiers change the completion for a position in general.

## chdir

[`$chdir(<directory>)`](https://rsteube.github.io/carapace/carapace/action/chDir.html) changes the directory.

```yml
["$chdir(/tmp)", "$(pwd)"]
```

## list

[`$list(<delimiter>)`](https://rsteube.github.io/carapace/carapace/action/actionMultiParts.html) completes values as unique list with given delimiter.

```yml
["$list(,)", "a", "b", "c", "d"]
```

## multiparts

[`$multiparts(<delimiter>)`](https://rsteube.github.io/carapace/carapace/invokedAction/toMultiPartsA.html) completes values splitted on given delimiter separately.

```yml
["$multiparts(/)", "a", "a/b", "a/c", "b", "b/a"]
```

## nospace

`$nospace` prevents space suffix being added to the inserted values.

```yml
["$nospace", "one", "two"]
```

## noflag

`$noflag` disables flag parsing for the corresponding (sub)command.

```yml
["$noflag"]
```
