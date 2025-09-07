package jobserver

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"teleport-jobworker/pkg/job"
)

// StartRequest defines the Start request body.
type StartRequest struct {
	Program string   `json:"program"`
	Args    []string `json:"args"`
}

// StartResponse defines the Start response body.
type StartResponse struct {
	ID    string  `json:"id"`
	Error *string `json:"error"`
}

// StopResponse defines the Stop response body.
type StopResponse struct {
	ID    string  `json:"id"`
	Error *string `json:"error"`
}

// StatusResponse defines the GetStatus response body.
type StatusResponse struct {
	ID       string  `json:"id"`
	Status   string  `json:"status"`
	ExitCode *int    `json:"exitCode"`
	Error    *string `json:"error"`
}

// OutputResponse defines the GetOutput response body.
type OutputResponse struct {
	ID     string  `json:"id"`
	Stdout string  `json:"stdout"`
	Stderr string  `json:"stderr"`
	Error  *string `json:"error"`
}

// ErrorResponse defines error response body for status codes: 401, 404, 500.
type ErrorResponse struct {
	Error string `json:"error"`
}

// responseJSON prepares the response body as JSON.
func responseJSON(w http.ResponseWriter, payload any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		log.Printf("json.Encoder.Encode() failed to encode response: %v", err)
	}
}

// responseError prepares the error response body as JSON.
func responseError(w http.ResponseWriter, err error) {
	// 404 Not Found, when a job is not in Manager's job table
	if errors.Is(err, job.ErrNotFound) {
		responseJSON(w, ErrorResponse{err.Error()}, http.StatusNotFound)
		return
	}
	// 500 Internal Server Error, for any issues with internal job functions
	responseJSON(w, ErrorResponse{err.Error()}, http.StatusInternalServerError)
}

// startHandler handles HTTPS requests to POST /jobs/start
func (s *Server) startHandler(w http.ResponseWriter, r *http.Request) {
	var startRequest StartRequest
	if err := json.NewDecoder(r.Body).Decode(&startRequest); err != nil {
		responseJSON(w, ErrorResponse{err.Error()}, http.StatusBadRequest)
		return
	}

	jobID := s.manager.Start(r.Context(), startRequest.Program, startRequest.Args)

	responseJSON(w, StartResponse{ID: jobID}, http.StatusCreated)
}

// stopHandler handles HTTPS reqs to POST /jobs/{id}/stop
func (s *Server) stopHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := s.manager.Stop(r.Context(), id)
	if err != nil {
		responseError(w, err)
		return
	}

	responseJSON(w, StopResponse{ID: id}, http.StatusOK)
}

// getStatusHandler handles HTTPS reqs to GET /jobs/{id}
func (s *Server) getStatusHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	status, err := s.manager.GetStatus(r.Context(), id)
	if err != nil {
		responseError(w, err)
		return
	}

	responseJSON(w, StatusResponse{
		ID:       id,
		Status:   status.State,
		ExitCode: status.ExitCode,
	}, http.StatusOK)
}

// getOutputHandler handles HTTPS reqs to GET /jobs/{id}/output
func (s *Server) getOutputHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	stdout, stderr, err := s.manager.GetOutput(r.Context(), id)
	if err != nil {
		responseError(w, err)
		return
	}

	responseJSON(w, OutputResponse{
		ID:     id,
		Stdout: stdout,
		Stderr: stderr,
	}, http.StatusOK)
}
