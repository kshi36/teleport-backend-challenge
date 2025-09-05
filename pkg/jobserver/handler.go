package jobserver

import (
	"encoding/json"
	"net/http"
	"teleport-jobworker/pkg/job"
)

// StartReq defines the Start request body.
type StartReq struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

// SimpleRes defines the Start and Stop response body.
type SimpleRes struct {
	ID    string  `json:"id"`
	Error *string `json:"error"`
}

// StatusRes defines the GetStatus response body.
type StatusRes struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	ExitCode *int    `json:"exitCode"`
	Error    *string `json:"error"`
}

// OutputRes defines the GetOutput response body.
type OutputRes struct {
	ID     string  `json:"id"`
	Stdout string  `json:"stdout"`
	Stderr string  `json:"stderr"`
	Error  *string `json:"error"`
}

// ErrRes defines error response body for status codes: 401, 404, 500.
type ErrRes struct {
	Error string `json:"error"`
}

// resJSON prepares the response body as JSON.
func resJSON(w http.ResponseWriter, payload any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// resError prepares the error response body as JSON.
func resError(w http.ResponseWriter, err error) {
	// 404 Not Found
	if err == job.ErrNotFound {
		resJSON(w, ErrRes{err.Error()}, http.StatusNotFound)
		return
	}
	// 500 Internal Server Error
	resJSON(w, ErrRes{err.Error()}, http.StatusInternalServerError)
}

// startHandler handles HTTPS requests to POST /jobs/start
func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	var req StartReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resJSON(w, ErrRes{err.Error()}, http.StatusBadRequest)
		return
	}

	// Start call
	jobID := s.mgr.Start(r.Context(), req.Program, req.Args)

	resJSON(w, SimpleRes{ID: jobID}, http.StatusCreated)
}

// stopHandler handles HTTPS reqs to POST /jobs/{id}/stop
func (s *Server) stopHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// Stop call
	err := s.mgr.Stop(r.Context(), id)
	if err != nil {
		resError(w, err)
		return
	}

	resJSON(w, SimpleRes{ID: id}, http.StatusOK)
}

// getStatusHandler handles HTTPS reqs to GET /jobs/{id}
func (s *Server) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// GetStatus call
	status, err := s.mgr.GetStatus(r.Context(), id)
	if err != nil {
		resError(w, err)
		return
	}

	resJSON(w, StatusRes{
		ID:       id,
		Status:   status.State,
		ExitCode: status.ExitCode,
	}, http.StatusOK)
}

// getOutputHandler handles HTTPS reqs to GET /jobs/{id}/output
func (s *Server) getOutputHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	// GetOutput call
	stdout, stderr, err := s.mgr.GetOutput(r.Context(), id)
	if err != nil {
		resError(w, err)
		return
	}

	resJSON(w, OutputRes{
		ID:     id,
		Stdout: stdout,
		Stderr: stderr,
	}, http.StatusOK)
}
