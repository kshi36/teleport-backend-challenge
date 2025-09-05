package main

import (
	"crypto/tls"
	"log"
	"net/http"

	"teleport-jobworker/pkg/job"
	"teleport-jobworker/pkg/jobserver"
)

func main() {
	// for now, just use localhost, port 8443
	addr := ":8443"

	// create new Manager
	mgr := job.NewManager()

	// create job Server with mux
	js := jobserver.NewServer(mgr)
	server := &http.Server{
		Addr:    addr,
		Handler: js,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13, // enforce TLS 1.3
		},
	}
	log.Fatal(server.ListenAndServeTLS("certs/cert.pem", "certs/key.pem"))
}
