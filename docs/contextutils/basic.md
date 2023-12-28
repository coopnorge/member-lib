# Basic usage

## Functions

### 1. AddKeyValue

AddKeyValue adds a key to the context and assigned to it a value.

- #### Parameters

  - context (context.Context): The context to which a key-value pair will
be added.
  - key (any): The key to be added to the specified context.
  - value (any): The value associated with the key being added to
the context.

- #### Returns

  - context: The context containing the added key-value pair.

- #### Example

```go
package main

import (
  ...
)

func main() {
  ctx := AddKeyValue(context.Background(), "foo", "bar")

  fmt.Println(ctx.Value("foo")) // Output: bar
}
```

### 2. GetKeyValue

GetKeyValue relies on GoLang generics to retrieve the value of a specific key
in the context.

- #### Parameters

  - context (context.Context): The context from which to retrieve a key value.
  - key (specified type using Golang generics): The key for which we want to
retrieve the associated value.

- #### Returns

  - value: The value we retrieve from the context.
  - error: An error, if any, encountered during the value retrieval from
context.

- #### Example

```go
package main

import (
  ...
)

func main() {
  ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, "foo", "bar")

  v, err := GetKeyValue[string, string](ctxWithValue, "foo")
  if err != nil {
      log.Fatal(err)
  }

  fmt.Println(v) // Output: bar
}
```

### 3. RemoveKey

RemoveKey removes a key and its value from context.

- #### Parameters

  - context (context.Context): The context to which a key-value pair will be
removed.
  - key (any): The key to be removed in the specified context.

- #### Returns

  - context: The context following the removal of the key-value pair.

- #### Example

```go
package main

import (
  ...
)

func main() {
  ctx := context.Background()
	ctxWithValue := context.WithValue(ctx, "foo", "bar")

  ctxWithoutValue := RemoveKey(ctxWithValue, "foo", "bar")

  fmt.Println(ctxWithoutValue.Value("foo")) // Output: <nil>
}
```
