// Package coordinator help manage the primary function of your application or service in a cleaner way.
//
// This concept is inspired by the standard Game Loop
// architecture, which handles major tasks such as initialization, state
// management, calculations, and rendering.
//
// The `ServiceCoordinator` adopts a similar approach, allowing multiple
// core goroutines within your service to be managed effectively.
//
// This is achieved by delegating the control of the main loop and determining
// the operational logic for starting or stopping processes such as HTTP or gRPC
// servers. In addition, it can efficiently run auxiliary goroutines for tasks
// such as Datadog monitoring, finite state machine (FSM) management, or other
// elements depending on the specific use case.
package coordinator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/coopnorge/member-lib/coordinator/process"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
)

type (
	// ServiceCoordinator manages the lifecycle of an application.
	// Responsible for overseeing the processes and tasks that constitute the app,
	// and provides mechanisms for controlled starts, stops, and signal handling.
	ServiceCoordinator struct {
		// signals are the OS signals that ServiceCoordinator listens to. When one of these signals
		// is received, ServiceCoordinator will initiate its stopping procedures. Typical signals
		// might include SIGINT (Ctrl+C) or SIGTERM (termination request).
		signals []os.Signal
		// stopTimeout specifies the maximum duration ServiceCoordinator will wait when trying to
		// gracefully stop the application. If this timeout is exceeded, ServiceCoordinator will
		// forcefully terminate the app. This ensures that applications do not hang
		// indefinitely during shutdown.
		stopTimeout time.Duration
		// processes is a list of processes that ServiceCoordinator will run in the background
		// during the application's lifecycle. These tasks run concurrently with the main
		// application process and can be thought of as auxiliary services or routines
		// that support the primary functions of the application.
		processes []process.Process
		// mainContext is the primary context for the ServiceCoordinator. It governs the entire lifecycle
		// of the application and its associated processes. When this context is cancelled,
		// it signals all derived contexts to begin their shutdown procedures.
		mainContext context.Context
		// mainContextCancel is the associated cancel function for the mainContext. Invoking
		// this function will cancel the mainContext and begin the shutdown process for
		// ServiceCoordinator and its managed tasks.
		mainContextCancel func()
	}

	// Options sets of configurations for ServiceCoordinator.
	Options func(o *ServiceCoordinator)

	// coreContextKey child context key.
	contextOfServiceCoordinator struct {
		id string
	}
)

// SetForceStopTimeout redefines force shutdown timeout.
func SetForceStopTimeout(t time.Duration) Options {
	return func(c *ServiceCoordinator) { c.stopTimeout = t }
}

// AddProcesses that will be executed in background of main loop.
func AddProcesses(p ...process.Process) Options {
	return func(c *ServiceCoordinator) {
		c.processes = append(c.processes, p...)
	}
}

// NewServiceCoordinator instance to manage the application.
func NewServiceCoordinator(opts ...Options) (b *ServiceCoordinator) {
	b = &ServiceCoordinator{
		signals:     []os.Signal{syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT},
		stopTimeout: 60 * time.Second,
	}

	b.mainContext, b.mainContextCancel = context.WithCancel(context.Background())
	for _, o := range opts {
		o(b)
	}

	return
}

// Start initializes and runs the ServiceCoordinator's main loop
// It launches the goroutines / processes and monitors for interrupt signals.
func (c *ServiceCoordinator) Start() error {
	// The WaitGroup helps to synchronize sub goroutines / processes, ensuring
	// it completed before exiting the function.
	var bgTasksWG sync.WaitGroup
	// This channel listens for OS-level interrupt signals.
	interruptSignal := make(chan os.Signal, 1)
	// Create a context and its associated error group for the goroutines / processes.
	processErrorGroup, processErrorGroupCtx := errgroup.WithContext(c.mainContext)

	// Initialization of goroutines / processes.
	for _, p := range c.processes {
		proc := p // redefine the var within the scope of loop, so that each goroutine gets its own copy

		// Define a goroutines / processes termination workflow.
		processErrorGroup.Go(func() error {
			<-processErrorGroupCtx.Done()

			newUUID, errNewUUID := uuid.NewUUID()
			if errNewUUID != nil {
				return fmt.Errorf("unable to generate UUID for process, err: %w", errNewUUID)
			}

			procStopCtx, procStopCtxCancel := context.WithTimeout(context.Background(), c.stopTimeout)
			procStopCtx = context.WithValue(procStopCtx, contextOfServiceCoordinator{id: newUUID.String()}, proc.GetName())

			defer procStopCtxCancel()

			return proc.OnStop(procStopCtx) //nolint:contextcheck // false positive, extended by context.WithValue
		})

		bgTasksWG.Add(1)

		// Define the goroutines / processes execution logic.
		processErrorGroup.Go(func() error {
			defer bgTasksWG.Done()

			err := proc.OnStart(processErrorGroupCtx)
			if err == nil {
				return nil
			}

			if process.IsCriticalToStop(proc) {
				return fmt.Errorf("critical error on start of process %s: %w", proc.GetName(), err)
			}

			return nil
		})
	}

	// Register the ServiceCoordinator's signals to the interruptSignal channel.
	signal.Notify(interruptSignal, c.signals...)

	// Main loop monitors for interrupts or cancellations in the processErrorGroup's context.
	processErrorGroup.Go(func() error {
		for {
			select {
			case <-processErrorGroupCtx.Done():
				// Context is done, probably due to an error or forced shutdown.
				return processErrorGroupCtx.Err()
			case <-interruptSignal:
				// Interrupt received, begin the shutdown process.
				return c.Stop()
			}
		}
	})

	// Wait for all goroutines / processes  in the error group to complete.
	// If any error occurs, and it's not due to cancellation, it's returned.
	if err := processErrorGroup.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

// Stop in graceful mode and terminate all goroutines / processes.
func (c *ServiceCoordinator) Stop() error {
	if c.mainContextCancel != nil {
		c.mainContextCancel()
	}

	return nil
}
