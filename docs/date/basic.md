# Basic usage

## Functions

### 1. ParseDateComponent

ParseDateComponent parses year, month or day strings and returns a
time object.

- #### Parameters

  - layout (string): The layout specifying the format of the date component.
For example, "2006" or "06".
  - date component (string): The date component to be parsed. For example,
"2024".

- #### Returns

  - time.Time: A date object representing the parsed date component.
  - error: An error, if any, encountered during the parsing of the
date component string.

- #### Example

```go
package main

import (
  ...
)

func main() {
  date, err := ParseDateComponent("2006", "2024")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(date) // Output: 2024-01-01 00:00:00 +0000 UTC
}
```

### 2. ParseYear

ParseYear parses year string

- #### Parameters

  - year (string): The year string to be parsed.
  - layout (string): The layout specifying the format of the year string.
For example, "2006" or "06".

- #### Returns

  - year (uint16): A uint16 number representing the parsed year.
  - error: An error, if any, encountered during the parsing of the year
string.
  
- #### Example

```go
package main

import (
  ...
)

func main() {
  year, err := ParseYear("2024", "2006")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(year) // Output: 2024
}
```

### 3. ParseMonth

ParseMonth parses month string

- #### Parameters

  - month (string): The month string to be parsed.
  - layout (string): The layout specifying the format of the month string. For example,
"Jan", "January", "01" or "1"

- #### Returns

  - month (uint8): A uint8 number representing the parsed month.
  - error: An error, if any, encountered during the parsing of the month
string.
  
- #### Example

```go
package main

import (
  ...
)

func main() {
  month, err := ParseMonth("5", "1")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(month) // Output: 5
}
```

### 4. ParseDayOfTheMonth

ParseDayOfTheMonth parses day string

- #### Parameters

  - day (string): The day string to be parsed.
  - layout (string): The layout specifying the format of the day string.
For example, "2", "_2" or "02"

- #### Returns

  - day (uint8): A uint8 number representing the parsed day.
  - error: An error, if any, encountered during the parsing
of the day string.
  
- #### Example

```go
package main

import (
  ...
)

func main() {
  day, err := ParseDayOfTheMonth("15", "2")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(day) // Output: 15
}
```

### 5. IsLeapYear

IsLeapYear determines whether a specific year is a leap year.

- #### Parameters

  - year (string): The year to be checked for leap year status.
  - layout (string): The layout specifying the format of the year
string. For example, "2006" or "06".
  
- #### Returns

  - bool: A boolean indicating whether the specified year is a leap
year (true) or not (false).
  - error: An error, if any, encountered during the parsing of the
year string.

- #### Example

```go
package main

import (
  ...
)

func main() {
  leap, err := IsLeapYear("2024", "2006")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(leap) // Output: true
}
```
