package jobserver

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"teleport-jobworker/pkg/job"
	"testing"
	"testing/synctest"
)

const (
	user1token    = "user1_token"
	user2token    = "user2_token"
	fakeusertoken = "fakeuser_token"
)

// initTestServer spins up a test HTTPS API server and pre-generated dummy ID for testing.
func initTestServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()

	manager := job.NewManager()
	jobServer := NewServer(manager)

	var id string
	var err error
	synctest.Test(t, func(t *testing.T) {
		// pre-populate Manager with a dummy Job, ID to test endpoints
		ctx := job.WithUserInfo(context.Background(), "user1", job.User)
		id, err = manager.Start(ctx, "/bin/echo", []string{"hello world"})
		if err != nil {
			t.Errorf("(*Manager).Start() error: %s", err)
		}

		// wait for dummy Job to be in "completed" state
		synctest.Wait()
	})

	ts := httptest.NewUnstartedServer(jobServer)
	ts.TLS = &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	ts.StartTLS()

	t.Cleanup(func() {
		ts.Close()
	})

	return ts, id
}

func TestStartHandler(t *testing.T) {
	ts, _ := initTestServer(t)

	shortCmd := `{"program":"/bin/echo","args":["hello world"]}`
	request, _ := http.NewRequest("POST", ts.URL+"/jobs/start", bytes.NewBufferString(shortCmd))
	request.Header.Set("Authorization", "Bearer "+user1token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusCreated {
		t.Errorf("startHandler() expected %d, got %d", http.StatusCreated, response.StatusCode)
	}
	response.Body.Close()
}

func TestStopHandler(t *testing.T) {
	ts, id := initTestServer(t)

	request, _ := http.NewRequest("POST", ts.URL+"/jobs/"+id+"/stop", nil)
	request.Header.Set("Authorization", "Bearer "+user1token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("stopHandler() expected %d, got %d", http.StatusOK, response.StatusCode)
	}
	response.Body.Close()
}

func TestStatusHandler(t *testing.T) {
	ts, id := initTestServer(t)

	request, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id, nil)
	request.Header.Set("Authorization", "Bearer "+user1token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusOK, response.StatusCode)
	}
	defer response.Body.Close()

	var statusResponse StatusResponse
	err = json.NewDecoder(response.Body).Decode(&statusResponse)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if statusResponse.Status != job.Completed {
		t.Errorf("GetStatus() expected completed, got %v", statusResponse.Status)
	}
}

func TestOutputHandler(t *testing.T) {
	ts, id := initTestServer(t)

	request, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id+"/output", nil)
	request.Header.Set("Authorization", "Bearer "+user1token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusOK {
		t.Errorf("getOutputHandler() expected %d, got %d", http.StatusOK, response.StatusCode)
	}
	response.Body.Close()
}

func TestJobNotFound(t *testing.T) {
	ts, _ := initTestServer(t)

	// GetStatus request, with fake id of a job that doesn't exist
	request, _ := http.NewRequest("GET", ts.URL+"/jobs/fake_id", nil)
	request.Header.Set("Authorization", "Bearer "+user1token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusNotFound, response.StatusCode)
	}
	defer response.Body.Close()

	var statusResponse StatusResponse
	err = json.NewDecoder(response.Body).Decode(&statusResponse)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if *statusResponse.Error != job.ErrNotFound.Error() {
		t.Errorf("GetStatus() expected error: %v, got %v", job.ErrNotFound.Error(), *statusResponse.Error)
	}
}

func TestWrongUserNotFound(t *testing.T) {
	ts, id := initTestServer(t)

	// GetStatus request, with wrong token (user2token)
	request, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id, nil)
	request.Header.Set("Authorization", "Bearer "+user2token)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusNotFound {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusNotFound, response.StatusCode)
	}
	defer response.Body.Close()

	var statusResponse StatusResponse
	err = json.NewDecoder(response.Body).Decode(&statusResponse)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if *statusResponse.Error != job.ErrNotFound.Error() {
		t.Errorf("GetStatus() expected error: %v, got %v", job.ErrNotFound.Error(), *statusResponse.Error)
	}
}

func TestUnauthorized(t *testing.T) {
	ts, id := initTestServer(t)

	request, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id, nil)
	request.Header.Set("Authorization", "Bearer "+fakeusertoken)
	request.Header.Set("Content-Type", "application/json")

	response, err := ts.Client().Do(request)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("startHandler() expected %d, got %d", http.StatusUnauthorized, response.StatusCode)
	}
	response.Body.Close()
}
