package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"teleport-jobworker/pkg/job"
	"teleport-jobworker/pkg/jobserver"
)

func main() {
	addr := jobserver.DefaultHost

	// create new Manager to inject into job Server
	manager := job.NewManager()

	// create job Server with mux to use with HTTPS
	jobServer := jobserver.NewServer(manager)
	server := &http.Server{
		Addr:    addr,
		Handler: jobServer,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}
	log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
}
