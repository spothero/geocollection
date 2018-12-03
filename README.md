# Location-Based Collection for Go

[![GoDoc](https://godoc.org/github.com/spothero/geocollection?status.svg)](https://godoc.org/github.com/spothero/geocollection)
[![Build Status](https://travis-ci.org/spothero/geocollection.png?branch=master)](https://travis-ci.org/spothero/geocollection)
[![codecov](https://codecov.io/gh/spothero/geocollection/branch/master/graph/badge.svg)](https://codecov.io/gh/spothero/geocollection)
[![Go Report Card](https://goreportcard.com/badge/github.com/spothero/geocollection)](https://goreportcard.com/report/github.com/spothero/geocollection)

geocollection builds on top of [Google's S2 Library](https://github.com/golang/geo) to provide a location-based collection for Go.
With this library you can store data based on latitude and longitude. Once it is stored, you may
retrieve items within a specific distance by supplying another latitude and longitude.

The library also exposes fine-grained controls over the covering parameters used by S2 to provide
maximum performance tuning.

API documentation and examples can be found in the [GoDoc](https://godoc.org/github.com/spothero/geocollection).

## License
Apache 2
