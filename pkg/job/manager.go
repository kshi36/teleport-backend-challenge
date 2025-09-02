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

// Start creates a job and assigns a unique job ID.
func (m *Manager) Start(program string, args []string) string {
	newJob := NewJob(program, args)

	m.mutex.Lock()
	m.jobs[newJob.ID] = newJob
	m.mutex.Unlock()

	go newJob.Run()

	return newJob.ID
}

// Stop kills the job of specified job ID
func (m *Manager) Stop(jobID string) error {
	// retrieve job
	m.mutex.RLock()
	job, ok := m.jobs[jobID]
	m.mutex.RUnlock()

	if !ok {
		return ErrJobNotFound
	}

	return job.Stop()
}

// GetStatus queries the job ID and returns job status, exit code, error
func (m *Manager) GetStatus(jobID string) (JobStatus, error) {
	// retrieve job
	m.mutex.RLock()
	job, ok := m.jobs[jobID]
	m.mutex.RUnlock()

	if !ok {
		return JobStatus{}, ErrJobNotFound
	}

	return job.Status(), nil
}

// GetOutput queries the job ID and returns stdout, stderr.
func (m *Manager) GetOutput(jobID string) (stdout, stderr string, err error) {
	// retrieve job
	m.mutex.RLock()
	job, ok := m.jobs[jobID]
	m.mutex.RUnlock()

	if !ok {
		return "", "", ErrJobNotFound
	}

	return job.Output()
}
