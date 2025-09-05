// Package job implements functions to interact with jobs,
// such as start, stop, query status, and get output of jobs.
package job

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"sync"

	"github.com/google/uuid"
)

// Job lifecycle
const (
	Starting  = "starting"
	Running   = "running"
	Failed    = "failed"
	Completed = "completed"
	Stopped   = "stopped"
)

var ErrNotFound = errors.New("job not found")

// Job is a Linux process started by the service.
type Job struct {
	ID     string
	cmd    *exec.Cmd
	outBuf safeBuffer
	errBuf safeBuffer

	status      JobStatus
	statusMutex sync.RWMutex
}

// JobStatus holds job status information.
type JobStatus struct {
	State    string
	ExitCode *int
}

// safeBuffer provides concurrent r/w access to buffer content.
// This allows safe access to process's stdout/stderr while job is running.
type safeBuffer struct {
	buf      bytes.Buffer
	bufMutex sync.RWMutex
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.bufMutex.Lock()
	defer s.bufMutex.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.bufMutex.RLock()
	defer s.bufMutex.RUnlock()
	return s.buf.String()
}

// newJob creates a new Job struct with state "Starting".
func newJob(program string, args []string) *Job {
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

// run forks a new process and manages job lifecycle.
func (j *Job) run() {
	// start the process
	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	err := j.cmd.Start()
	// unsuccessful starting the process
	if err != nil {
		j.status = JobStatus{State: Failed}
		return
	}

	// successful starting the process
	j.status = JobStatus{State: Running}

	// wait for process completion
	go j.wait()
}

// wait sits on the process until completion, then updates state.
func (j *Job) wait() {
	j.cmd.Wait()

	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	exitCode := j.cmd.ProcessState.ExitCode()

	// process terminated by signal (state "Stopped")
	if exitCode == -1 {
		j.status = JobStatus{State: Stopped, ExitCode: &exitCode}
		return
	}

	// process completed
	j.status = JobStatus{State: Completed, ExitCode: &exitCode}
}

// stop kills the job process with signal SIGKILL.
func (j *Job) stop() error {
	j.statusMutex.Lock()
	defer j.statusMutex.Unlock()

	// job is not currently running, graceful return
	if j.status.State != Running && j.status.State != Starting {
		return nil
	}

	// job process still starting, coalesce into ErrNotFound
	if j.cmd.Process == nil {
		return ErrNotFound
	}

	// send SIGKILL signal
	err := j.cmd.Process.Kill()
	if err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	return nil
}

// getStatus returns the job's status and exit code.
func (j *Job) getStatus() JobStatus {
	j.statusMutex.RLock()
	defer j.statusMutex.RUnlock()

	return j.status
}

// getOutput returns the job's stdout/stderr data.
func (j *Job) getOutput() (stdout, stderr string) {
	return j.outBuf.String(), j.errBuf.String()
}
