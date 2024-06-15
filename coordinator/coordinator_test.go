package coordinator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/coopnorge/member-lib/coordinator/process"
	"github.com/stretchr/testify/assert"
)

type stubProcess struct {
	mu sync.Mutex

	started chan bool
	stopped chan bool

	severity         process.Severity
	isFailingOnStart bool
}

func (m *stubProcess) GetSeverity() process.Severity {
	return m.severity
}

func (m *stubProcess) OnStart(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isFailingOnStart {
		return errors.New("have error OnStart always - error flag is set")
	}

	m.started <- true

	return nil
}

func (m *stubProcess) OnStop(_ context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopped <- true

	return nil
}

func (m *stubProcess) GetName() string {
	return "UnitTestStubProcess"
}

func TestServiceCoordinator(t *testing.T) {
	majorUnitTestProcessStartSignal := make(chan bool)
	majorUnitTestProcessStopSignal := make(chan bool)
	majorUnitTestProcess := &stubProcess{
		severity: process.TaskSeverityMajor,
		started:  majorUnitTestProcessStartSignal,
		stopped:  majorUnitTestProcessStopSignal,
	}

	minorUnitTestProcessStartSignal := make(chan bool)
	minorUnitTestProcessStopSignal := make(chan bool)
	minorUnitTestProcess := &stubProcess{
		severity:         process.TaskSeverityMajor,
		started:          minorUnitTestProcessStartSignal,
		stopped:          minorUnitTestProcessStopSignal,
		isFailingOnStart: true,
	}

	sc := NewServiceCoordinator(
		AddProcesses(majorUnitTestProcess, minorUnitTestProcess),
		SetForceStopTimeout(time.Millisecond),
	)

	assert.Len(t, sc.processes, 2, "Expected 2 processes")
	assert.True(t, sc.stopTimeout == time.Millisecond)

	go func() {
		_ = sc.Start()
	}()

	select {
	case <-majorUnitTestProcessStartSignal:
		// Pass
	case <-time.After(time.Millisecond):
		t.Error("Start was not called in the expected timeframe")
	}

	select {
	case <-minorUnitTestProcessStartSignal:
		t.Error("Start was not expected to be called in the expected timeframe")
	case <-time.After(time.Millisecond):
		// Pass
	}

	time.Sleep(500 * time.Millisecond)

	stopErr := sc.Stop()
	assert.NoError(t, stopErr, "Expected no error, got %v", stopErr)

	select {
	case <-majorUnitTestProcessStopSignal:
		// Pass
	case <-time.After(time.Millisecond):
		t.Error("Stop was not called in the expected timeframe")
	}

	select {
	case <-minorUnitTestProcessStopSignal:
		// Pass
	case <-time.After(time.Millisecond):
		t.Error("Stop was not called in the expected timeframe")
	}
}

func exampleNewHTTPServer() *stubProcess {
	return new(stubProcess)
}
func exampleNewPrometheusMetrics() *stubProcess {
	return new(stubProcess)
}

func Example_newServiceCoordinator() {
	// By assuming you need to wrap you application core startup logic in one handled algorithm.
	// Things like how to start and stop server, monitoring, metrics in controllable goroutines.
	// Then you can simplify your code base by delighting it to `ServiceCoordinator`.
	//
	// Before creating new ServiceCoordinator
	// You may need to define configuration, like reading env vars or file
	// Then you can create new instance of ServiceCoordinator
	applicationForceStopTimeout := time.Minute
	setOfOptions := []Options{
		SetForceStopTimeout(applicationForceStopTimeout),
		// each of processes must be implementation of type Process interface from `process` package.
		AddProcesses(exampleNewHTTPServer(), exampleNewPrometheusMetrics()),
	}

	// Now pass your setup to service coordinator
	sc := NewServiceCoordinator(setOfOptions...)

	// And you can start it, inside OS signals will be handled, so you don't need to write graceful shutdown here.
	if err := sc.Start(); err != nil {
		// Do you fancy code here, this is just example
		fmt.Printf("unable to start my application: %v", err)
		os.Exit(1)
	}
}
