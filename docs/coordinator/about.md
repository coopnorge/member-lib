# What is coordinator component

Coordinator component help manage the primary function of your application
or service in a cleaner way. This concept is inspired by the standard Game Loop
architecture, which handles major tasks such as initialization, state
management, calculations, and rendering.

The `ServiceCoordinator` adopts a similar approach, allowing multiple
core goroutines within your service to be managed effectively.

This is achieved by delegating the control of the main loop and determining
the operational logic for starting or stopping processes such as HTTP or gRPC
servers. In addition, it can efficiently run auxiliary goroutines for tasks
such as Datadog monitoring, finite state machine (FSM) management, or other
elements depending on the specific use case.
