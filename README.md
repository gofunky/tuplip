# tuplip

[![GoDoc](https://godoc.org/github.com/gofunky/tuplip?status.svg)](https://godoc.org/github.com/gofunky/tuplip)
[![Go Report Card](https://goreportcard.com/badge/github.com/gofunky/tuplip)](https://goreportcard.com/report/github.com/gofunky/tuplip)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/b2ba6ca3e18c48c799ebdfa3962b9e81)](https://www.codacy.com/app/gofunky/tuplip?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=gofunky/tuplip&amp;utm_campaign=Badge_Grade)
[![Dependabot Status](https://api.dependabot.com/badges/status?host=github&repo=gofunky/tuplip)](https://dependabot.com)
[![GitHub License](https://img.shields.io/github/license/gofunky/tuplip.svg)](https://github.com/gofunky/tuplip/blob/master/LICENSE)
[![GitHub last commit](https://img.shields.io/github/last-commit/gofunky/tuplip.svg)](https://github.com/gofunky/tuplip/commits/master)
[![Microbadger Version](https://images.microbadger.com/badges/version/gofunky/tuplip.svg)](https://microbadger.com/images/gofunky/tuplip "Docker Version")
[![Microbadger Layers](https://images.microbadger.com/badges/image/gofunky/tuplip.svg)](https://microbadger.com/images/gofunky/tuplip "Docker Layers")
[![Docker Pulls](https://img.shields.io/docker/pulls/gofunky/tuplip.svg)](https://hub.docker.com/r/gofunky/tuplip)

generate a complete set of Docker tags that depict all tuples of dependency versions

## Installation

### Using `go get`

```bash
go get -u github.com/gofunky/tuplip
```

### Using Docker

```bash
docker pull gofunky/tuplip
```

## Using tulip

### Using the binary

```bash
echo "dep _:1.0.0" | tuplip exec
```

### Using Docker

```bash
echo "dep _:1.0.0" | docker run --rm -i gofunky/tuplip exec
```

## Standard Input

Separate the input tags either by newlines or by spaces.

### Unversioned Aliases

Unversioned input tags define dependencies without versions or special build parameters that define a separate output portion.

#### Example

```bash
echo "something fancy" | tuplip exec
```

#### Result

```bash
something
fancy
fancy-something
```

### Versioned Aliases

Versioned input tags define dependencies with versions.
Then, the version is altered and depicted in all its variants.

#### Example

```bash
echo "go:1.2.3" | tuplip exec
```

#### Result

```bash
go
go1
go1.2
go1.2.3
```

### Versioned Wildcard Aliases

A versioned wildcard input tag is used to depict the different version representation of the project itself.

#### Example

```bash
echo "_:1.0 dep" | tuplip exec
```

#### Result

```bash
1
1.0
dep
1-dep
1.0-dep
```

## Parameters

### excludeMajor

`excludeMajor` excludes the major versions (e.g., `go1` for `go:1.2.3`) from the result set.

#### Example

```bash
echo "go:1.2.3" | tuplip exec excludeMajor
```

#### Result

```bash
go
go1.2
go1.2.3
```

### excludeMinor

`excludeMinor` excludes the minor versions (e.g., `go1.2` for `go:1.2.3`) from the result set.

#### Example

```bash
echo "go:1.2.3" | tuplip exec excludeMinor
```

#### Result

```bash
go
go1
go1.2.3
```

### excludeBase

`excludeBase` excludes the base alias (e.g., `go` for `go:1.2.3`) from the result set.

#### Example

```bash
echo "go:1.2.3" | tuplip exec excludeBase
```

#### Result

```bash
go1
go1.2
go1.2.3
```

### addLatest

`addLatest` adds the `latest` tag to the result set.

#### Example

```bash
echo "_:1.2.3" | tuplip exec addLatest
```

#### Result

```bash
1
1.2
1.2.3
latest
```
