package jobserver

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// For the prototype, pre-generated Bearer tokens will be assigned to specific userIDs.
// In the future, tokens will be auto-generated and assigned from the server,
// and stored securely.
var userTokens = map[string]string{ // userID -> token
	"user1":  "user1_token",
	"user2":  "user2_token",
	"admin1": "admin1_token",
}

const (
	DefaultBaseURL            = "https://localhost:8443"
	messageJobStarted         = "Job started with ID %s\n"
	messageJobStopped         = "Job stopped for ID %s\n"
	messageJobStatus          = "Job status for ID %s\nStatus: %s\nExit code: %s\n"
	messageJobOutput          = "Job output for ID %s\nstdout:\n%s\nstderr:\n%s\n"
	messageJobNotFound        = "Job not found for ID %s\n"
	messageUnauthorizedAction = "User is not authorized to perform the action\n"
	messageInternalError      = "Jobserver internal error\n"
)

// Client sends job management requests to the HTTPS API server.
type Client struct {
	*http.Client
	URL string
}

// NewClient configures an HTTP Client for communication with the job Server.
func NewClient(url string) (*Client, error) {
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

	if url == "" {
		url = DefaultBaseURL
	}
	return &Client{Client: client, URL: url}, nil
}

// StartJob creates an HTTP request and parses response for the /jobs/start endpoint.
func (c *Client) StartJob(user, program string, args []string) (string, error) {
	var requestBuf bytes.Buffer
	requestBody := StartRequest{
		Program: program,
		Args:    args,
	}
	if err := json.NewEncoder(&requestBuf).Encode(requestBody); err != nil {
		return "", err
	}

	request, err := http.NewRequest("POST", c.URL+"/jobs/start", &requestBuf)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+userTokens[user])
	request.Header.Set("Content-Type", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusCreated:
		var startResponse StartResponse
		err = json.NewDecoder(response.Body).Decode(&startResponse)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(messageJobStarted, startResponse.ID), nil
	case http.StatusUnauthorized:
		return messageUnauthorizedAction, nil
	default: // coalesce "BadRequest" into "InternalServerError"
		return messageInternalError, nil
	}
}

// StopJob creates an HTTP request and parses response for the /jobs/{id}/stop endpoint.
func (c *Client) StopJob(user, jobID string) (string, error) {
	request, err := http.NewRequest("POST", c.URL+"/jobs/"+jobID+"/stop", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+userTokens[user])
	request.Header.Set("Content-Type", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		var stopResponse StopResponse
		err = json.NewDecoder(response.Body).Decode(&stopResponse)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf(messageJobStopped, stopResponse.ID), nil
	case http.StatusNotFound:
		return fmt.Sprintf(messageJobNotFound, jobID), nil
	case http.StatusUnauthorized:
		return messageUnauthorizedAction, nil
	default:
		return messageInternalError, nil
	}
}

// GetJobStatus creates an HTTP request and parses response for the /jobs/{id} endpoint.
func (c *Client) GetJobStatus(user, jobID string) (string, error) {
	request, err := http.NewRequest("GET", c.URL+"/jobs/"+jobID, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+userTokens[user])
	request.Header.Set("Content-Type", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		var statusResponse StatusResponse
		err = json.NewDecoder(response.Body).Decode(&statusResponse)
		if err != nil {
			return "", err
		}

		exitCode := ""
		if statusResponse.ExitCode != nil {
			exitCode = strconv.Itoa(*statusResponse.ExitCode)
		}
		return fmt.Sprintf(messageJobStatus,
			statusResponse.ID, statusResponse.Status, exitCode), nil
	case http.StatusNotFound:
		return fmt.Sprintf(messageJobNotFound, jobID), nil
	case http.StatusUnauthorized:
		return messageUnauthorizedAction, nil
	default:
		return messageInternalError, nil
	}
}

// GetJobOutput creates an HTTP request and parses response for the /jobs/{id}/output endpoint.
func (c *Client) GetJobOutput(user, jobID string) (string, error) {
	request, err := http.NewRequest("GET", c.URL+"/jobs/"+jobID+"/output", nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+userTokens[user])
	request.Header.Set("Content-Type", "application/json")

	response, err := c.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	switch response.StatusCode {
	case http.StatusOK:
		var outputResponse OutputResponse
		err = json.NewDecoder(response.Body).Decode(&outputResponse)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf(messageJobOutput,
			outputResponse.ID, outputResponse.Stdout, outputResponse.Stderr), nil
	case http.StatusNotFound:
		return fmt.Sprintf(messageJobNotFound, jobID), nil
	case http.StatusUnauthorized:
		return messageUnauthorizedAction, nil
	default:
		return messageInternalError, nil
	}
}
