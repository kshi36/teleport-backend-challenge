package jobserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
)

const DefaultBaseURL = "https://localhost:8443"

// Client sends job management requests to the HTTPS API server.
type Client struct {
	client *http.Client
	url    string
}

// NewClient configures an HTTP Client for communication with the job Server.
func NewClient() (*Client, error) {
	// configure job Client to trust self-signed TLS certificate
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(cert); !ok {
		return nil, fmt.Errorf("failed to append cert to CertPool")
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certPool,
			},
		}}

	return &Client{client, DefaultBaseURL}, nil
}

// userToToken simply appends a "_token" string to a userID to generate a Bearer token.
// In the future, tokens will be auto-generated (eg. JWT) according to user info.
func userToToken(user string) string {
	return user + "_token"
}

// StartJob creates an HTTP request and parses response for the /jobs/start endpoint.
func (c *Client) StartJob(user, program string, args []string) (*StartResponse, error) {
	var requestBuf bytes.Buffer
	requestBody := StartRequest{
		Program: program,
		Args:    args,
	}
	if err := json.NewEncoder(&requestBuf).Encode(requestBody); err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", c.url+"/jobs/start", &requestBuf)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+userToToken(user))
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var startResponse StartResponse
	err = json.NewDecoder(response.Body).Decode(&startResponse)
	if err != nil {
		return nil, err
	}
	return &startResponse, nil
}

// StopJob creates an HTTP request and parses response for the /jobs/{id}/stop endpoint.
func (c *Client) StopJob(user, jobID string) (*StopResponse, error) {
	request, err := http.NewRequest("POST", c.url+"/jobs/"+jobID+"/stop", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+userToToken(user))
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var stopResponse StopResponse
	err = json.NewDecoder(response.Body).Decode(&stopResponse)
	if err != nil {
		return nil, err
	}
	return &stopResponse, nil
}

// GetJobStatus creates an HTTP request and parses response for the /jobs/{id} endpoint.
func (c *Client) GetJobStatus(user, jobID string) (*StatusResponse, error) {
	request, err := http.NewRequest("GET", c.url+"/jobs/"+jobID, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+userToToken(user))
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var statusResponse StatusResponse
	err = json.NewDecoder(response.Body).Decode(&statusResponse)
	if err != nil {
		return nil, err
	}
	return &statusResponse, nil
}

// GetJobOutput creates an HTTP request and parses response for the /jobs/{id}/output endpoint.
func (c *Client) GetJobOutput(user, jobID string) (*OutputResponse, error) {
	request, err := http.NewRequest("GET", c.url+"/jobs/"+jobID+"/output", nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+userToToken(user))
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var outputResponse OutputResponse
	err = json.NewDecoder(response.Body).Decode(&outputResponse)
	if err != nil {
		return nil, err
	}
	return &outputResponse, nil
}
