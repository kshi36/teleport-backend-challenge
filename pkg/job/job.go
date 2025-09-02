package job

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/google/uuid"
)

const (
	Starting  = "starting"
	Running   = "running"
	Failed    = "failed"
	Completed = "completed"
	Stopped   = "stopped"
)

// Job is a Linux process started by the service.
type Job struct {
	ID     string
	cmd    *exec.Cmd
	outBuf bytes.Buffer
	errBuf bytes.Buffer

	status      JobStatus
	statusMutex sync.RWMutex
}

// JobStatus holds job status information.
type JobStatus struct {
	State    string
	ExitCode int
	Error    string
}

// New creates a new Job struct with state "Starting".
func NewJob(program string, args []string) *Job {
	ID := uuid.NewString()

	job := Job{
		ID:     ID,
		cmd:    exec.Command(program, args...),
		status: JobStatus{State: Starting},
	}

	// attach stdout and stderr buffers
	job.cmd.Stdout = &job.outBuf
	job.cmd.Stderr = &job.errBuf

	return &job
}

// Run forks a new process and manages job lifecycle.
func (j *Job) Run() {
	// start the process
	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	err := j.cmd.Start()
	// unsuccessful starting the process
	if err != nil {
		j.status = JobStatus{
			State:    Failed,
			ExitCode: -1,
			Error:    fmt.Sprintf("process failed to start: %v", err),
		}
		return
	}

	// successful starting the process
	j.status = JobStatus{State: Running}

	// wait for process completion
	go j.wait()
}

// wait sits on the process until completion, then updates state.
func (j *Job) wait() {
	err := j.cmd.Wait()

	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	exitCode := j.cmd.ProcessState.ExitCode()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// job was stopped by SIGKILL (Stopped)
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				if ws.Signaled() && ws.Signal() == syscall.SIGKILL {
					j.status = JobStatus{
						State:    Stopped,
						ExitCode: -1,
						Error:    "process killed (SIGKILL)",
					}
					return
				}
			}
			// job exits with non-zero exit code (Completed)
			j.status = JobStatus{
				State:    Completed,
				ExitCode: exitCode,
				Error:    fmt.Sprintf("process stopped with exit code %d", exitCode),
			}
			return
		}
		// job process quit unexpectedly
		j.status = JobStatus{
			State:    Failed,
			ExitCode: -1,
			Error:    fmt.Sprintf("unexpected process error: %v", err),
		}
		return
	}

	// job exits with exit code 0 (Completed)
	j.status = JobStatus{
		State:    Completed,
		ExitCode: exitCode,
	}
}

// Stop kills the job process with signal SIGKILL.
func (j *Job) Stop() error {
	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	// job is not currently running, graceful return
	if j.status.State != Running && j.status.State != Starting {
		return nil
	}

	// job process still starting
	if j.cmd.Process == nil {
		return ErrJobProcessNull
	}

	// send SIGKILL signal
	err := j.cmd.Process.Kill()
	if err != nil && err != os.ErrProcessDone {
		return err
	}

	return nil
}

// Status returns the job's status, exit code, and errors
func (j *Job) Status() JobStatus {
	j.statusMutex.RLock()
	defer j.statusMutex.RUnlock()

	return j.status
}

// Output returns the job's stdout/stderr data
func (j *Job) Output() (stdout, stderr string, err error) {
	j.statusMutex.RLock()
	defer j.statusMutex.RUnlock()

	// only return data when process is completed
	// in the future, data should be streamed and retrievable while running
	if j.status.State != Completed {
		return "", "", ErrJobNotCompleted
	}

	return j.outBuf.String(), j.errBuf.String(), nil
}
