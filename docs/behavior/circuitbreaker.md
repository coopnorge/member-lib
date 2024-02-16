# Circuit Breaker

Package `breaker`  implements the Circuit Breaker pattern.

This pattern prevents an application from performing operations
that are likely to fail, allowing it to maintain high performance
and availability even when some parts of a system are not functioning
optimally. The breaker package provides a simple yet flexible way to
integrate circuit breaking logic into your applications.

## Code Example

```go
package main

import (
  "log"

  "github.com/coopnorge/member-lib/behavior/breaker"
)

func main() {
  cbCfg := &breaker.Configuration{
    MaxFailuresThreshold: "100", // Amount of allowed failures before state will be Open.
    ResetTimeout:         "1",   // Time duration used to switch state from Open.
  }

  cb, cbErr := breaker.NewCircuitBreaker(cbCfg)
  if cbErr != nil {
    // TODO Deal with that
  }

  var myActionThatINeedExecuteInCircuitBreaker breaker.Action
  myActionThatINeedExecuteInCircuitBreaker = func() (any, error) {
    return "Hey ho, let's go", nil
  }

  result, resultErr := cb.Proceed(myActionThatINeedExecuteInCircuitBreaker)
  if resultErr != nil {
    // TODO Deal with that again
  }

  log.Println(result)

  // What if you need check state or reset Circuit Breaker?
  if cb.GetState().IsState(breaker.StateOpen){
    cb.Reset() // Ok, now state of Circuit Breaker in initial.
  }
}
```
