<!-- markdownlint-disable-file MD031 -->
<!-- markdownlint-disable-file MD010 -->

# About

The Object Pool design pattern is a way to manage the allocation and reuse of
objects, especially when creating new instances of these objects is costly in
terms of resources or performance.

[More about pattern](https://en.wikipedia.org/wiki/Object_pool_pattern)

## When to use `ResourcePoolManger`

> Package: creational/pool/pool.go

### Some possible use cases

- You might need to control your `connections` to something, then you could use
  this component to control state of connection and reuse it between
  executions.
- `Messages` - maybe you have need to track send messages, for example RPC
  connection. Then there is option to set amount of usages and Resource Pool
  Manger will create and destroy resourced.

### How schematically it's works

```text
     ┌─────────────────────────────────────────────┐
     │ Your Project Code                           │     ┌──────────────────────────┐
     │                                             │     │ Configure                │
     │                                             │     │                          │
     │   ┌──────────────────┐                      │     │ Pool Size and Usage Count│
     │   │DatabasedConnector│                      │     └─────────────┬────────────┘
     │   └┬─────────────────┘                      │                   │
     │    │                                        │                   │
     │   ┌▼──────────────────────────┬┐            │     ┌─────────────▼───────────┐
     │   │Database Connection Factory│┼────────────┼─────┤► Resource Pool Manager  │
     │   └───────────────────────────┴┘            │     └──────┬┬─────────────────┤
     │                                             │            ││                 │
     │                                             │            ││                 │
     │                                             │ ┌──────────┤►Acquire Resource │
     │   ┌──────────────────────────┐              │ │          ││                 │
     │   │Repository                │              │ │          ││                 │
     │   ├──────────┬┬──────────────┘              │ │       ┌──┤►Release Resource │
     │   │ Find User││                             │ │       │  └──────────────────┘
     │   └──────┬───┼┘                             │ │       │
     │          │   │  1.Ask to get free Connector │ │       │
     │          │   └──────────────────────────────┼─┘       │
     │          │                                  │         │
     │          │      2.Return back Connector     │         │
     │          └──────────────────────────────────┼─────────┘
     │                                             │
     └─────────────────────────────────────────────┘
```

### Code Example

Add to your code Resource Pool Manager

```go
// Remote procedure call service - "RPC"
// `Connector` is a type for generic that must be managed in Resource Pool
rpc.connectorPool = creational.NewResourcePoolManger[Connector](
uint8(maxConn),
uint8(maxMsgPerConn),
&ConnectorFactory{cfg: cfg},
)
```

Simple usage:

```go
resourceErr := rpc.connectorPool.AcquireAndReleaseResource(func (connector *Connector) error {
return connector.Connect()
})

if resourceErr != nil {
return fmt.Errorf("unable to open connection to database, error: %w", resourceErr)
}
```

Multi error handling:

```go
var connErr error
resourceErr := rpc.connectorPool.AcquireAndReleaseResource(func (connector *Connector) error {
connErr := connector.Connect()
return connErr
})

if connErr != nill {
return fmt.Errorf("unable to open connection to database, error: %w", connErr)
}
if resourceErr != nil {
return fmt.Errorf("unable to acquire database connector, error: %w", resourceErr)
}
```

You also have control when to acquire and release resource

```go
connector, ackErr := rpc.connectorPool.AcquireResource()
if ackErr != nil {
return fmt.Errorf("unable to acquire database connector, error: %w", ackErr)
}
defer rpc.connectorPool.ReleaseResource(connector)

connector.DoMyWork()
// Do more fancy code ...
// Or maybe you don't need to do defer, so you could call it here: rpc.connectorPool.ReleaseResource(connector)

```
