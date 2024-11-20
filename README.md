# Member-Lib

[![Go Reference](https://pkg.go.dev/badge/github.com/coopnorge/member-lib/pkg.svg)](https://pkg.go.dev/github.com/coopnorge/member-lib/pkg)
[![CI](https://github.com/coopnorge/member-lib/pkg/actions/workflows/golang-ci.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/pkg/actions/workflows/golang-ci.yaml)
[![Security](https://github.com/coopnorge/member-lib/pkg/actions/workflows/security-scan.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/pkg/actions/workflows/security-scan.yaml)

## Overview

`member-lib` is a Go library by Coop Norge designed to provide reusable
packages for various use cases within the Member domain.

It aims to enhance productivity and provide generalized, abstract solutions
applicable across a variety of use cases.

## Installing

To install the library, use:

```bash
go get github.com/coopnorge/member-lib/pkg/{package_of_interest}
```

### Usage

Import the desired package in your Go project:

```go
import "github.com/coopnorge/member-lib/pkg/<package_name>"
```

#### Example

```go
package main

import (
  "fmt"
  "github.com/coopnorge/member-lib/pkg/package1"
)

func main() {
  result := package1.Function()
  fmt.Println(result)
}
```

## Contribution

We welcome contributions!

For code style we have own preferences but not so different for standard Go
style. However, good to keep in mind:

- Abstract and not contain domain related names.
- Components must be covered with tests.
- Code style go standards with some Coop preferences.
  if you are internally contributor, please
  visit: [Coop Golan Guidelines](https://inventory.internal.coop/docs/default/component/guidelines/languages/go)

#### Please follow these steps:

1. Fork the repository.
2. Create a new branch.
3. Make your changes and commit them.
4. Push your changes and create a pull request.

## Testing

To verify if technical documentation is built properly run:

```shell
docker compose up techdocs
```

### Golang

Components must be:

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
