<!-- markdownlint-disable-file MD009 -->

# String package

It simplifies the conversion (casting) between strings and various numeric
types in Go, such as int, int64, uint, etc., in a secure manner that gracefully
handles type overflows.

Simple use case:

```go
package main

import (
  "fmt"

  "github.com/coopnorge/member-lib/strconv"
)

func main() {
  given := "-32767"
  parsed, parserErr := strconv.ToWholeNumber[int8](given)
  if parserErr != nil {
    // TODO Deal with that
  } else {
    fmt.Printf("my number now is type of int => %v", parsed)
  }
}

```
