package job

import (
	"strings"
	"testing"
	"time"
)

var shortCmd = []string{"/bin/echo", "hello world"}
var longCmd = []string{"/bin/sleep", "2"}
var invalidCmd = []string{"/invalid/cmd", "I am invalid"}
var exit1Cmd = []string{"sh", "-c", "exit 1"}

func TestStartShort(t *testing.T) {
	m := NewManager()
	jobID := m.Start(shortCmd[0], shortCmd[1:])

	time.Sleep(50 * time.Millisecond)

	//get status
	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Completed || status.ExitCode != 0 {
		t.Errorf("GetStatus() expected completed w/ exit code 0, got %v, code %v",
			status.State, status.ExitCode)
	}

	// get output
	stdout, _, err := m.GetOutput(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if !strings.Contains(stdout, "hello world") {
		t.Errorf("Unexpected stdout: %q", stdout)
	}
}

func TestStartLong(t *testing.T) {
	m := NewManager()
	jobID := m.Start(longCmd[0], longCmd[1:])

	time.Sleep(50 * time.Millisecond)

	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	// get status: running
	if status.State != Running {
		t.Errorf("GetStatus() expected running, got %v", status.State)
	}

	time.Sleep(3 * time.Second)

	// get status: completed
	status, err = m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if status.State != Completed || status.ExitCode != 0 {
		t.Errorf("GetStatus() expected completed w/ exit code 0, got %v, code %v",
			status.State, status.ExitCode)
	}
}

func TestStartInvalidCmd(t *testing.T) {
	m := NewManager()
	jobID := m.Start(invalidCmd[0], invalidCmd[1:])

	time.Sleep(50 * time.Millisecond)

	//get status
	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Failed {
		t.Errorf("GetStatus() expected failed , got %v", status.State)
	}
}

func TestNonZeroExit(t *testing.T) {
	m := NewManager()
	jobID := m.Start(exit1Cmd[0], exit1Cmd[1:])

	time.Sleep(50 * time.Millisecond)

	//get status
	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Completed || status.ExitCode != 1 {
		t.Errorf("GetStatus() expected completed w/ exit code 1, got %v, code %v",
			status.State, status.ExitCode)
	}
}

func TestStop(t *testing.T) {
	m := NewManager()
	jobID := m.Start(longCmd[0], longCmd[1:])

	time.Sleep(50 * time.Millisecond)

	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Running {
		t.Errorf("GetStatus() expected running, got %v", status.State)
	}

	// perform Stop
	err = m.Stop(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	time.Sleep(50 * time.Millisecond)

	status, err = m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if status.State != Stopped {
		t.Errorf("GetStatus() expected stopped, got %v", status.State)
	}
}

func TestStopAfterCompleted(t *testing.T) {
	m := NewManager()
	jobID := m.Start(longCmd[0], longCmd[1:])

	time.Sleep(3 * time.Second)

	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Completed {
		t.Errorf("GetStatus() expected completed, got %v", status.State)
	}

	// perform Stop after complete
	err = m.Stop(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	status, err = m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	if status.State != Completed {
		t.Errorf("GetStatus() expected stopped, got %v", status.State)
	}
}

func TestOutputBeforeCompleted(t *testing.T) {
	m := NewManager()
	jobID := m.Start(longCmd[0], longCmd[1:])

	time.Sleep(50 * time.Millisecond)

	// get output before completed, should result in error
	_, _, err := m.GetOutput(jobID)
	if err != ErrJobNotCompleted {
		t.Errorf("process should not have output, not completed")
	}

	time.Sleep(3 * time.Second)
	status, err := m.GetStatus(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}

	if status.State != Completed {
		t.Errorf("GetStatus() expected completed, got %v", status.State)
	}

	_, _, err = m.GetOutput(jobID)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
}
