# casm

CASM:  building blocks for distributed applications

[![Godoc Reference](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/lthibault/casm)

[![Go Report Card](https://goreportcard.com/badge/github.com/lthibault/casm?style=flat-square)](https://goreportcard.com/report/github.com/lthibault/casm)

## Installation

You first need to build the CaptnProto transport encodings.  First install the `capnp` tool (`brew install capnp` on macOS).

Next, install the Go bindings for CaptnProto:

```bash
make deps
```

Finally, build the encodings:

```bash
make
```
