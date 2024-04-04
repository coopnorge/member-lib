# Phone number

Package provides light code to validate phone number in string representation.

Example:

```go
package main

import (
  "fmt"

  "github.com/coopnorge/member-lib/validator/phone"
)

func main() {
  numbers := []string{"+12345678901", "+19995553333", "12345", "+1"}

  for _, number := range numbers {
    if phone.Validate(phone.FormatE164, number) {
      fmt.Printf("Number %s is valid E.164 format.\n", number)
    } else {
      fmt.Printf("Number %s is NOT valid E.164 format.\n", number)
    }
  }
}
```

