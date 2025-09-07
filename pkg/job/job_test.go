package job

import (
	"strings"
	"testing"
	"testing/synctest"
)

var shortCmd = []string{"/bin/echo", "hello world"}
var longCmd = []string{"/bin/sleep", "2"}

func TestRun(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		job := newJob(shortCmd[0], shortCmd[1:])
		job.run()

		status := job.getStatus()
		if status.State != Running {
			t.Errorf("getStatus() expected running with exit code 0, got %v, code %v",
				status.State, *status.ExitCode)
		}

		// wait for goroutine of (*Job).wait() to complete
		synctest.Wait()

		status = job.getStatus()
		if status.State != Completed || *status.ExitCode != 0 {
			t.Errorf("getStatus() expected completed with exit code 0, got %v, code %v",
				status.State, *status.ExitCode)
		}

		stdout, _ := job.getOutput()
		if !strings.Contains(stdout, "hello world") {
			t.Errorf("Unexpected stdout: %q", stdout)
		}
	})
}

func TestStop(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		job := newJob(longCmd[0], longCmd[1:])
		job.run()

		status := job.getStatus()
		if status.State != Running {
			t.Errorf("getStatus() expected running with exit code 0, got %v, code %v",
				status.State, *status.ExitCode)
		}

		err := job.stop()
		if err != nil {
			t.Errorf("stop() error: %s", err.Error())
		}

		synctest.Wait()

		status = job.getStatus()
		if status.State != Stopped || *status.ExitCode != -1 {
			t.Errorf("getStatus() expected stopped with exit code -1, got %v, code %v",
				status.State, *status.ExitCode)
		}
	})
}

func TestStopAfterCompleted(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		job := newJob(longCmd[0], longCmd[1:])
		job.run()

		status := job.getStatus()
		if status.State != Running {
			t.Errorf("getStatus() expected running with exit code 0, got %v, code %v",
				status.State, *status.ExitCode)
		}

		synctest.Wait()

		// stop after job completed should not be an error, graceful return
		err := job.stop()
		if err != nil {
			t.Errorf("stop() error: %s", err.Error())
		}

		status = job.getStatus()
		if status.State != Completed || *status.ExitCode != 0 {
			t.Errorf("getStatus() expected completed with exit code 0, got %v, code %v",
				status.State, *status.ExitCode)
		}
	})
}

func TestStartInvalidCmd(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		m, ctx := initManagerContext(Admin)
		jobID := m.Start(ctx, invalidCmd[0], invalidCmd[1:])

		synctest.Wait()

		status, err := m.GetStatus(ctx, jobID)
		if err != nil {
			t.Errorf("GetStatus() error: %s", err.Error())
		}

		if status.State != Failed {
			t.Errorf("GetStatus() expected failed , got %v", status.State)
		}
	})
}
