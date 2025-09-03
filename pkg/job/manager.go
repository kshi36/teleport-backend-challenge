package job

import (
	"sync"
)

// Manager tracks every job created by the service.
type Manager struct {
	mutex sync.RWMutex
	jobs  map[string]*Job
}

// NewManager creates a new Manager with empty job table.
func NewManager() *Manager {
	return &Manager{
		jobs: map[string]*Job{},
	}
}

// readJob retrieves a Job if the jobID exists in the table.
func (m *Manager) readJob(jobID string) *Job {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	return m.jobs[jobID]
}

// Start creates a job and assigns a unique job ID.
func (m *Manager) Start(program string, args []string) string {
	newJob := newJob(program, args)

	m.mutex.Lock()
	m.jobs[newJob.ID] = newJob
	m.mutex.Unlock()

	go newJob.run()

	return newJob.ID
}

// Stop kills the job of specified job ID
func (m *Manager) Stop(jobID string) error {
	job := m.readJob(jobID)
	if job == nil {
		return ErrNotFound
	}

	return job.stop()
}

// GetStatus queries the job ID and returns job status, exit code, error
func (m *Manager) GetStatus(jobID string) (JobStatus, error) {
	job := m.readJob(jobID)
	if job == nil {
		return JobStatus{}, ErrNotFound
	}

	return job.getStatus(), nil
}

// GetOutput queries the job ID and returns stdout, stderr.
func (m *Manager) GetOutput(jobID string) (stdout, stderr string, err error) {
	job := m.readJob(jobID)
	if job == nil {
		return "", "", ErrNotFound
	}

	stdout, stderr = job.getOutput()
	return stdout, stderr, nil
}
