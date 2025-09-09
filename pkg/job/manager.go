package job

import (
	"context"
	"errors"
	"sync"
)

var ErrNotFound = errors.New("job not found")
var ErrUnauthorized = errors.New("unauthorized action, no user provided")

// Manager roles
const (
	User  = "user"
	Admin = "admin"
)

// Manager tracks every job created by the service.
type Manager struct {
	mutex sync.RWMutex
	jobs  map[string]*jobRecord // jobID -> (userID, Job)
}

// jobRecord tracks user ID associated to Job.
type jobRecord struct {
	userID string
	job    *Job
}

// Context includes user ID and role for use with Manager functions.
type userContextKey struct{}
type userInfo struct {
	id   string
	role string
}

// NewManager creates a new Manager with empty job table.
func NewManager() *Manager {
	return &Manager{
		jobs: map[string]*jobRecord{},
	}
}

// Start creates a job and assigns a unique job ID.
func (m *Manager) Start(ctx context.Context, program string, args []string) (string, error) {
	userID, _, ok := getUserInfo(ctx)
	if !ok {
		return "", ErrUnauthorized
	}

	newJob := newJob(program, args)

	m.mutex.Lock()
	m.jobs[newJob.ID] = &jobRecord{job: newJob, userID: userID}
	m.mutex.Unlock()

	go newJob.run()

	return newJob.ID, nil
}

// Stop kills the job of specified job ID.
func (m *Manager) Stop(ctx context.Context, jobID string) error {
	job, err := m.readJob(ctx, jobID)
	if err != nil {
		return err
	}

	return job.stop()
}

// GetStatus queries the job ID and returns job status, exit code.
func (m *Manager) GetStatus(ctx context.Context, jobID string) (JobStatus, error) {
	job, err := m.readJob(ctx, jobID)
	if err != nil {
		return JobStatus{}, err
	}

	return job.getStatus(), nil
}

// GetOutput queries the job ID and returns stdout, stderr.
func (m *Manager) GetOutput(ctx context.Context, jobID string) (stdout, stderr string, err error) {
	job, err := m.readJob(ctx, jobID)
	if err != nil {
		return "", "", err
	}

	stdout, stderr = job.getOutput()
	return stdout, stderr, nil
}

// readJob retrieves a Job if the jobID exists in table and user has valid role.
func (m *Manager) readJob(ctx context.Context, jobID string) (*Job, error) {
	userID, role, ok := getUserInfo(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	m.mutex.RLock()
	record := m.jobs[jobID]
	m.mutex.RUnlock()

	// job is not found in table
	// and, user can only act on owned jobs, unless they have admin role
	if record == nil || (role != Admin && userID != record.userID) {
		return nil, ErrNotFound
	}

	return record.job, nil
}

// WithUserInfo adds user information data (userID and role) into the context.
func WithUserInfo(ctx context.Context, id, role string) context.Context {
	return context.WithValue(ctx, userContextKey{}, &userInfo{id, role})
}

// getUserInfo retrieves user information data (userID and role) from context.
func getUserInfo(ctx context.Context) (userID, role string, ok bool) {
	userInfo, ok := ctx.Value(userContextKey{}).(*userInfo)
	return userInfo.id, userInfo.role, ok
}
