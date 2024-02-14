# Member-Lib

[![CI](https://github.com/coopnorge/member-lib/actions/workflows/golang-ci.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/actions/workflows/golang-ci.yaml)
[![CI](https://github.com/coopnorge/member-lib/actions/workflows/techdocs.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/actions/workflows/techdocs.yaml)
[![Security](https://github.com/coopnorge/member-lib/actions/workflows/security-scan.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/actions/workflows/security-scan.yaml)

This repository contains a collection of Go packages designed from Member
domain to provide generalized, abstract solutions applicable across a variety
of use cases.

## Installing

### Installing specific package

```bash
go get github.com/coopnorge/member-lib/{package_of_interest}
```

### Importing in code

```go
package yourpackage

import (
  "github.com/coopnorge/member-lib/{package_of_interest}"
)

func YourCode() {
  // TODO More of your code...
}
```

## Contribution

### Documentation

Technical documentation stored in `docs`, each sub folder represents golang
component.
To check how technical documentation loose like before merging it, you can run
it in docker.

```shell
docker compose up techdocs
```

### Golang

Components must be:

- Abstract and not contain domain related names.
- Components must be covered with tests.
- Code style
  follow [Coop Golan Guidelines](https://inventory.internal.coop/docs/default/component/guidelines/languages/go/#http-client)

For validating and building Golang code you can use Coops docker tool.

Build images:

```shell
docker compose build
```

See available targets:

```shell
docker compose run --rm golang-devtools help
```

Validate golang:

```shell
docker compose run --rm golang-devtools validate VERBOSE=all
```

Test golang code:

```shell
docker compose run --rm golang-devtools test VERBOSE=all
```
