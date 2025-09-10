package jobserver

import (
	"strings"
	"teleport-jobworker/pkg/job"
	"testing"
)

func TestStartJob(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{ts.Client(), ts.URL}

	response, err := client.StartJob("user1", "/bin/echo", []string{"hello world"})
	if err != nil {
		t.Errorf("StartJob() error: %s", err.Error())
	}

	if response.Error != nil {
		t.Errorf("StartJob() job error: %s", *response.Error)
	}
}

func TestGetJobStatus(t *testing.T) {
	ts, id := initTestServer(t)
	client := &Client{ts.Client(), ts.URL}

	response, err := client.GetJobStatus("user1", id)
	if err != nil {
		t.Errorf("GetJobStatus() error: %s", err.Error())
	}

	if response.Error != nil {
		t.Errorf("GetJobStatus() job error: %s", *response.Error)
	}
}

func TestGetJobStatusNotFound(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{ts.Client(), ts.URL}

	response, err := client.GetJobStatus("user1", "fake_id")
	if err != nil {
		t.Errorf("GetJobStatus() error: %s", err.Error())
	}

	if response.Error == nil || !strings.Contains(*response.Error, job.ErrNotFound.Error()) {
		t.Errorf("GetJobStatus() expected %s, got %s", job.ErrNotFound.Error(), *response.Error)
	}
}

func TestStartJobUnauthorized(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{ts.Client(), ts.URL}

	response, err := client.StartJob("fakeuser", "/bin/echo", []string{"hello world"})
	if err != nil {
		t.Errorf("StartJob() error: %s", err.Error())
	}

	if response.Error == nil || !strings.Contains(*response.Error, ErrBadAuthentication) {
		t.Errorf("StartJob() expected %s, got %s", job.ErrUnauthorized.Error(), *response.Error)
	}
}
