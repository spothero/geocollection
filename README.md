# Location-Based Collection for Go

[![GoDoc](https://godoc.org/github.com/spothero/geocollection?status.svg)](https://godoc.org/github.com/spothero/geocollection)
[![Build Status](https://circleci.com/gh/spothero/geocollection.svg?style=shield)](https://circleci.com/gh/spothero/geocollection)
[![codecov](https://codecov.io/gh/spothero/geocollection/branch/master/graph/badge.svg)](https://codecov.io/gh/spothero/geocollection)
[![Go Report Card](https://goreportcard.com/badge/github.com/spothero/geocollection)](https://goreportcard.com/report/github.com/spothero/geocollection)

geocollection builds on top of [Google's S2 Library](https://github.com/golang/geo) to provide a location-based collection for Go.
With this library you can store data based on latitude and longitude. Once it is stored, you may
retrieve items within a specific distance by supplying another latitude and longitude.

The library also exposes fine-grained controls over the covering parameters used by S2 to provide
maximum performance tuning.

API documentation and examples can be found in the [GoDoc](https://godoc.org/github.com/spothero/geocollection).

## Linting

Run the linter using the command `make lint`.

A common linting error is the `fieldalignment` warning from the `govet` analyzer. `fieldalignment` errors arise when the order of a struct’s fields could be arranged differently to optimize the amount of allocated memory.

Imagine the following struct:
```
type MyObject struct {
    myBool   bool
    myString string
}
```

Running the linter would produce this output:
```
>> make lint
golangci-lint run

main.go:16:15: fieldalignment: struct with 16 pointer bytes could be 8 (govet)
type MyObject struct {
              ^
make: *** [lint] Error 1
```

The struct is more optimally arranged as:
```
type MyObject struct {
    myString string
    myBool   bool
}
```

A `fieldalignment` command line tool exists to help optimally arrange all the structs in a given file or package. Note that this tool will remove all existing comments within any structs it rearranges. Be sure to manually re-add any deleted comments after running the command.

Installation:
```
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
```

Utilization:
```
fieldalignment -fix {PATH_TO_FILE_OR_PACKAGE}
```

## License
Apache 2
