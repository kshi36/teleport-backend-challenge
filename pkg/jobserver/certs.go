package jobserver

import (
	"crypto/tls"
	_ "embed"
)

// The self-signed TLS certificate will be embedded into the binary.
// In the future, the job Server will use certificate issued by a Trusted CA.

//go:embed certs/cert.pem
var cert []byte

//go:embed certs/key.pem
var key []byte

// LoadTLSCertificate loads the self-signed TLS certificate for the job Server.
func LoadTLSCertificate() (tls.Certificate, error) {
	return tls.X509KeyPair(cert, key)
}
