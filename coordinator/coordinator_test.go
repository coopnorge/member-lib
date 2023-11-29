package coordinator

import (
	"context"
	"errors"
	"os"
	"sync"
	"syscall"
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

	time.Sleep(time.Millisecond)

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

func TestHandleShutdownSignals(t *testing.T) {
	coordinator := ServiceCoordinator{}

	callbackCalled := make(chan bool, 1)

	coordinator.HandleShutdownSignals(func(err error) {
		if err != nil {
			t.Errorf("Not expected error, but got %v", err)
		}

		callbackCalled <- true
	})

	time.AfterFunc(time.Microsecond, func() {
		p, _ := os.FindProcess(os.Getpid())
		_ = p.Signal(syscall.SIGTERM)
	})

	select {
	case <-callbackCalled:
		t.Fatal("HandleShutdownSignals returned error and it was error callback was called ")
	case <-time.After(time.Millisecond):
		// Pass
	}
}
