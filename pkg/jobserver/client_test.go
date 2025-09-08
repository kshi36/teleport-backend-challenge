package jobserver

import (
	"strings"
	"testing"
)

const (
	jobStarted       = "Job started"
	jobStatus        = "Job status"
	jobNotFound      = "Job not found"
	userUnauthorized = "User is not authorized"
)

func TestStartJob(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{Client: ts.Client(), URL: ts.URL}

	message, err := client.StartJob("user1", "/bin/echo", []string{"hello world"})
	if err != nil {
		t.Errorf("StartJob() error: %s", err.Error())
	}

	if !strings.Contains(message, jobStarted) {
		t.Errorf("StartJob() expected %s, got %s", jobStarted, message)
	}
}

func TestGetJobStatus(t *testing.T) {
	ts, id := initTestServer(t)
	client := &Client{Client: ts.Client(), URL: ts.URL}

	message, err := client.GetJobStatus("user1", id)
	if err != nil {
		t.Errorf("GetJobStatus() error: %s", err.Error())
	}

	if !strings.Contains(message, jobStatus) {
		t.Errorf("GetJobStatus() expected %s, got %s", jobStatus, message)
	}
}

func TestGetJobStatusNotFound(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{Client: ts.Client(), URL: ts.URL}

	message, err := client.GetJobStatus("user1", "fake_id")
	if err != nil {
		t.Errorf("GetJobStatus() error: %s", err.Error())
	}

	if !strings.Contains(message, jobNotFound) {
		t.Errorf("GetJobStatus() expected %s, got %s", jobNotFound, message)
	}
}

func TestStartJobUnauthorized(t *testing.T) {
	ts, _ := initTestServer(t)
	client := &Client{Client: ts.Client(), URL: ts.URL}

	message, err := client.StartJob("fakeuser", "/bin/echo", []string{"hello world"})
	if err != nil {
		t.Errorf("StartJob() error: %s", err.Error())
	}

	if !strings.Contains(message, userUnauthorized) {
		t.Errorf("StartJob() expected %s, got %s", jobStarted, message)
	}
}
