# Basic usage

An example of usage `coordinator` package in your entry-point function for the
application `main`, you may use this code snippet to adapt it to your own
needs.

```go
package main

import (
  ...
)

func main() {
  // ...
  // Before creating new ServiceCoordinator
  // You may need to define configuration, like reading env vars or file
  // Then you can create new instance of ServiceCoordinator
  forceStopTimout := 30 * time.Second
  sc := NewServiceCoordinator(forceStopTimout)

  // Add your processes to ServiceCoordinator that must run
  // It can be server handlers, monitoring or anything that must be used as goroutine.
  sc.AddProcess(NewDatadogProcess())
  sc.AddProcess(NewHTTPServer())

  // Start it
  if err := sc.Start(); err != nil {
    // Do you fancy code here, this is just example
    fmt.Printf("unable to start my application: %v", err)
    os.Exit(1)
  }

  // Handle OS signals to gracefully shut down processes.
  onShutdownSignalError := func(err error) {
    // Check the returned error channel from HandleSignals and act accordingly.
    // Do something else if you need if gracefully shutdown failed.

    if err != nil {
      fmt.Printf("failed gracefully stop application: %v", err)
      os.Exit(1)
    }
  }

  // Delegate your callback to handle errors on shutdown.
  sc.HandleShutdownSignals(onShutdownSignalError)
}
```
