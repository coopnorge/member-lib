<!-- markdownlint-disable-file MD009 -->
<!-- markdownlint-disable-file MD031 -->
# OpenAPI

Package provides utilities that help work with
[OpenAPI](https://swagger.io/specification/) generated code.

## Example

To simplify response handling and avoid writing logic to parse default
structure of OpenAPI might look like:

```go
type FooBarResponse struct {
  Body         []byte
  HTTPResponse *http.Response
  JSON200      *GetCardResponse
  JSON400      *ValidationProblemDetails
  JSON424      *ProblemDetails
  JSON500      *ProblemDetails
}
```

You can simply use generic
function `ExtractResponse[T any](resp *http.Response)` from this
package `net/openapi`.

Example of code call:
```go
package api

import "github.com/coopnorge/member-lib/net/openapi"

type ContractResponse struct {
  Name string `json:"name,omitempty"`
}

func DoingMyJob() {
  response, responseErr := anotherClient.GetMyContract(ctx, user) // Assume that response is FooBarResponse
  if responseErr != nil {
    // TODO I'll handle it
  }

  extractedResponse, extractedBadRequest, extractErr := openapi.ExtractResponse[ContractResponse](&Response{HTTPResponse: response, HTTPResponseBody: &response.Body})
  // TODO Handle extractErr
  // TODO Write code to work with extractedBadRequest or extractedResponse
}
```
