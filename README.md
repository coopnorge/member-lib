# Member-Lib

[![CI](https://github.com/coopnorge/member-lib/actions/workflows/ci.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/actions/workflows/ci.yaml)
[![Security](https://github.com/coopnorge/member-lib/actions/workflows/security-scan.yaml/badge.svg?branch=main)](https://github.com/coopnorge/member-lib/actions/workflows/security-scan.yaml)

This repository contains a collection of Go packages designed from Member
domain to provide generalized, abstract solutions applicable across a variety
of use cases.

## Installing

### In stalling specific package

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
