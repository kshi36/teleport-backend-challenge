// Package jobserver defines the HTTPS API handlers,
// including middleware for Bearer token authentication & authorization.
package jobserver

import (
	"net/http"
	"teleport-jobworker/pkg/job"
)

const DefaultHost = ":8443"

// Server provides an HTTP mux and job.Manager for job API calls.
type Server struct {
	mux     *http.ServeMux
	manager *job.Manager
}

// NewServer creates an HTTP mux with API endpoints for job functions.
func NewServer(manager *job.Manager) *Server {
	mux := http.NewServeMux()

	jobServer := &Server{
		mux:     mux,
		manager: manager,
	}

	mux.HandleFunc("POST /jobs/start", bearerAuth(jobServer.startHandler))
	mux.HandleFunc("POST /jobs/{id}/stop", bearerAuth(jobServer.stopHandler))
	mux.HandleFunc("GET /jobs/{id}/output", bearerAuth(jobServer.getOutputHandler))
	mux.HandleFunc("GET /jobs/{id}", bearerAuth(jobServer.getStatusHandler))

	return jobServer
}

// ServeHTTP allows the job Server to be used with http.Server.
func (js *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	js.mux.ServeHTTP(w, r)
}
