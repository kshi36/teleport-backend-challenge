package jobserver

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"teleport-jobworker/pkg/job"
	"testing"
	"time"
)

const (
	user1    = "user1"
	user2    = "user2"
	fakeuser = "fakeuser"
)

var shortCmd = `{"program":"/bin/echo","args":["hello world"]}`
var longCmd = `{"program":"/bin/sleep","args":["5"]}`

// initTestServer spins up a test HTTPS server with job Server mux
func initTestServer() *httptest.Server {
	mgr := job.NewManager()
	js := NewServer(mgr)

	// create HTTPS test server instance, with TLS 1.3
	ts := httptest.NewUnstartedServer(js)
	ts.TLS = &tls.Config{
		MinVersion: tls.VersionTLS13, // enforce TLS 1.3
	}
	ts.StartTLS()

	return ts
}

// makeStartReq makes an HTTP Start request, and parses response into jobID.
func makeStartReq(t *testing.T, ts *httptest.Server, cmd string) string {
	t.Helper()

	req, _ := http.NewRequest("POST", ts.URL+"/jobs/start", bytes.NewBufferString(cmd))
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("startHandler() expected %d, got %d", http.StatusCreated, res.StatusCode)
	}
	defer res.Body.Close()

	var simpleRes SimpleRes
	err = json.NewDecoder(res.Body).Decode(&simpleRes)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	return simpleRes.ID
}

func TestStartHandler(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start request
	req, _ := http.NewRequest("POST", ts.URL+"/jobs/start", bytes.NewBufferString(shortCmd))
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusCreated {
		t.Errorf("startHandler() expected %d, got %d", http.StatusCreated, res.StatusCode)
	}
	defer res.Body.Close()
}

func TestStopHandler(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start req
	id := makeStartReq(t, ts, longCmd)

	time.Sleep(200 * time.Millisecond)

	// Stop req
	req, _ := http.NewRequest("POST", ts.URL+"/jobs/"+id+"/stop", nil)
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("stopHandler() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("stopHandler() expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	defer res.Body.Close()
}

func TestStatusHandler(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start req
	id := makeStartReq(t, ts, shortCmd)

	time.Sleep(200 * time.Millisecond)

	// GetStatus req
	req, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	defer res.Body.Close()

	var statusRes StatusRes
	err = json.NewDecoder(res.Body).Decode(&statusRes)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if statusRes.Status != job.Completed {
		t.Errorf("GetStatus() expected completed, got %v", statusRes.Status)
	}
}

func TestOutputHandler(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start req
	id := makeStartReq(t, ts, shortCmd)

	time.Sleep(200 * time.Millisecond)

	// GetOutput req
	req, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id+"/output", nil)
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusOK {
		t.Errorf("getOutputHandler() expected %d, got %d", http.StatusOK, res.StatusCode)
	}
	defer res.Body.Close()
}

func TestNotFoundID(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// GetStatus req, w/ fake id
	req, _ := http.NewRequest("GET", ts.URL+"/jobs/fake_id", nil)
	req.Header.Set("Authorization", "Bearer "+user1)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusNotFound, res.StatusCode)
	}
	defer res.Body.Close()

	// Error field should be populated with job.ErrNotFound
	var statusRes StatusRes
	err = json.NewDecoder(res.Body).Decode(&statusRes)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if *statusRes.Error != job.ErrNotFound.Error() {
		t.Errorf("GetStatus() expected error: %v, got %v", job.ErrNotFound.Error(), *statusRes.Error)
	}
}

func TestNotFoundWrongUser(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start req
	id := makeStartReq(t, ts, shortCmd)

	time.Sleep(200 * time.Millisecond)

	// GetStatus req, w/ wrong user (user2)
	req, _ := http.NewRequest("GET", ts.URL+"/jobs/"+id, nil)
	req.Header.Set("Authorization", "Bearer "+user2)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("getStatusHandler() expected %d, got %d", http.StatusNotFound, res.StatusCode)
	}
	defer res.Body.Close()

	// Error field should be populated with job.ErrNotFound
	var statusRes StatusRes
	err = json.NewDecoder(res.Body).Decode(&statusRes)
	if err != nil {
		t.Errorf("JSON decoding error: %s", err.Error())
	}

	if *statusRes.Error != job.ErrNotFound.Error() {
		t.Errorf("GetStatus() expected error: %v, got %v", job.ErrNotFound.Error(), *statusRes.Error)
	}
}

func TestUnauthorized(t *testing.T) {
	ts := initTestServer()
	defer ts.Close()

	// Start req, unauthenticated (bad token)
	req, _ := http.NewRequest("POST", ts.URL+"/jobs/start", bytes.NewBufferString(shortCmd))
	req.Header.Set("Authorization", "Bearer "+fakeuser)
	req.Header.Set("Content-Type", "application/json")

	res, err := ts.Client().Do(req)
	if err != nil {
		t.Errorf("Do() error: %s", err.Error())
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("startHandler() expected %d, got %d", http.StatusUnauthorized, res.StatusCode)
	}
	defer res.Body.Close()
}
