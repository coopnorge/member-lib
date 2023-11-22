## Usage in Main

This acts as a guide for an example of usage `coordinator` package.
To simplify your logic in entry-point function for the application `main`, you
may use this code snippet to adapt it to your own needs.

```go
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
    fmt.Printf("unable to start ServiceCoordinator: %v", err)
    os.Exit(1)
  }
  
  // Handle OS signals to gracefully shut down processes.
  errChan := sc.HandleSignals()
  
  // You can do some other work here or just let the program wait for a shutdown signal.
  // ...
  
  // Check the returned error channel from HandleSignals and act accordingly.
  select {
  case err := <-errChan:
    if err != nil {
    fmt.Printf("There was an error stopping the services: %v", err)
    os.Exit(1)
  }
  }
}
```
