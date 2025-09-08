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

	cert, err := jobserver.LoadTLSCertificate()
	if err != nil {
		log.Fatal("failed to load TLS certificate")
	}

	server := &http.Server{
		Addr:    addr,
		Handler: jobServer,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{cert},
		},
	}
	log.Fatal(server.ListenAndServeTLS("", ""))
}
