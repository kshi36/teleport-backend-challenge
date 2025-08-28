# Job Worker Service \- Design Document

## Introduction

This design document describes a prototype job worker service, which exposes functionality to run arbitrary Linux programs as jobs. Users can start, stop, query status, and get the output of a job. Users access the service via a CLI that interacts with an API provided by an HTTPS server.

There are three primary components for the job worker service: library, API server, and CLI tool. The library implements the core functionality of the service, managing job lifecycles and tracking every job in the system. The API server wraps the library functionality via an HTTPS API, with endpoints for the job functions. The API endpoints will use Bearer token authentication, and all API communication will be secured over TLS. The service will follow a simple authorization scheme to allow regular users and admins different permissions for job management. The CLI tool provides a simple interface for users to make requests to the API server.

## Scope

The job worker service will have simplifications as a prototype.

* Jobs will run directly on the server machine on demand and remain in-memory. The service will not provide scheduling of jobs.  
* Completed jobs and outputs will be stored in-memory until the service is closed. There is no persistent storage of job logs, such as in an external database.  
* To simplify authentication, tokens will be pre-generated and mapped to user IDs. This determines if users can access the service functions. In addition, the HTTPS connection will use a self-signed TLS certificate, and the CLI client will be configured to trust this certificate explicitly. The TLS configuration will enforce TLS version 1.3 and use defaults from Go’s `crypto/tls` library for secure cipher suites.  
* The authorization scheme allows users to only operate on jobs started by them, while admins can operate on any job in the system. The prototype will not focus on an allow-list of commands users can run.  
* The CLI and API server will be designed to run on the same machine running the service (localhost).

## Library

The library contains the logic for all job management functions. A job will be represented as a Job struct with process information and states. A Manager struct will contain a global list of Jobs, and is responsible for starting, stopping, and querying jobs.

* The Job struct will contain information specific to one job, including metadata such as unique job ID, job owner, Linux program name, program arguments, PID, current status, exit code, and buffers for output. Job IDs will be generated as UUIDv4 via the `google/uuid` library. Job status transitions will be properly protected via synchronization.  
* The Manager struct maintains every job created by the service, with proper synchronization to enable concurrent users to perform job functions. It will contain a table mapping unique job IDs to Job structs, and also maintain the job IDs associated with each user ID for authorization.   
* A job lifecycle consists of the following states: Running, Completed, Failed, and Stopped. All statuses will include extra information when necessary.  
  * Running \- Job currently running.  
  * Completed \- Job successful, normal exit code and no errors.  
  * Failed \- Job failed. Includes varied error reports for abnormal exits, errors starting the job, or when an OS signal terminates the process.   
  * Stopped \- Job forcefully stopped by user.

Start(program, args) → jobID

* Start a new job with specified program path and program arguments.  
* Spawn the process, update status to Running, and create a goroutine to wait to capture output/errors and exit code.   
* On completion, goroutine will update status to either Completed or Failed.

Stop(jobID)

* Stop the job with specified job ID.  
* Send a SIGKILL signal to fully stop the process. Update status to Stopped.

GetStatus(jobID) → {status, exitCode}

* Query status of job with specified job ID.

GetOutput(jobID) → {outData, errData}

* Retrieve output and errors from stdout/stderr of job with specified job ID.

## API

The API server wraps the functionality of the job worker library. It contains endpoints to start, stop, query status, and get output of a job. The endpoint handlers will perform authentication and authorization checks for job requests. The endpoints will gracefully handle and report errors. For the prototype, API versioning is omitted for simplicity. Below is the proposed API with HTTP methods and simplified endpoints, where actual endpoints will be served over HTTPS. This includes notable headers, response codes, and JSON formats for requests and responses.

### Start a job

POST /jobs/start  
Authorization: Bearer \<token\>  
Content-Type: application/json  
Request body: {“program”: “/bin/sleep”, “args”: \[5\]}

201 Created → Job successfully started  
{“id”: “j-12345”, “error”: null}

401 Unauthorized → Missing or invalid Bearer token  
{“error”: “Error: bad authentication”}

404 Not Found → Program path not found  
{“error”: “Error: program not found”}

### Stop a job

POST /jobs/{id}/stop  
Authorization: Bearer \<token\>

200 OK → Job successfully stopped  
{“id”: “j-12345”, “error”: null}

401 Unauthorized → Missing or invalid Bearer token  
{“error”: “Error: bad authentication”}

403 Forbidden → User does not own the job ID, and is not admin  
{“error”: “Error: unauthorized action”}

404 Not Found → Job not found  
{“error”: “Error: job not found”}

409 Conflict → Job not running  
{“error”: “Error: job not running”}

### Get status of job

GET /jobs/{id}  
Authorization: Bearer \<token\>

200 OK → Job status retrieved  
{“id”: “j-12345”, “status”: “Running”, “exitCode”: null, “error”: null}

401 Unauthorized → Missing or invalid Bearer token  
{“error”: “Error: bad authentication”}

403 Forbidden → User does not own the job ID, and is not admin  
{“error”: “Error: unauthorized action”}

404 Not Found → Job not found  
{“error”: “Error: job not found”}

### Get output of job

GET /jobs/{id}/output

200 OK → Job output retrieved  
{“id”: “j-12345”, “stdout”: \[\], “stderr”: \[\], “error”: null}

401 Unauthorized → Missing or invalid Bearer token  
{“error”: “Error: bad authentication”}

403 Forbidden → User does not own the job ID, and is not admin  
{“error”: “Error: unauthorized action”}

404 Not Found → Job not found  
{“error”: “Error: job not found”}

## CLI

The CLI tool allows users to start, stop, query status, or get output of a job. These are requests sent to the API server. To run a Linux program, a user must provide the absolute path of the executable (eg. /bin/echo for echo program). The CLI will include an optional argument to specify the client with different user IDs for testing purposes (eg. \-u “user1”, \-u “admin”). The CLI will include basic input validation, such as for program not found, wrong usage, or wrong arguments.

### Client Examples

`jobctl start -- /bin/sleep 5`  
`Job started with ID j-12345.`

`jobctl status j-12345`  
`ID: j-12345`  
`Status: Running`  
`Exit code: n/a`

`jobctl stop j-12345`  
`Job with ID j-12345 stopped.`

`jobctl start -- /bin/echo “hello world”`  
`Job started with ID j-98765.`

`jobctl output j-98765`  
`Output:`  
`hello world`

## Security

The TLS configuration for the server will follow defaults from Go’s `crypto/tls` library to provide modern secure configurations. For the prototype, TLS version will be enforced to TLS 1.3 for strongest security. In addition, the API server will use a self-signed TLS certificate to implement the HTTPS connection. The certificate and private key pair will be pre-generated. In production, the server will use certificates issued by trusted Certificate Authorities to improve security and trust between clients and the server.

### Authentication

The job worker service will use HTTP Bearer token authentication to establish a trust layer between clients and servers. Users will hold access tokens, and the server will authenticate users by validating those tokens. Only users with valid tokens can perform job functions.

### Authorization

A simple authorization scheme will cover the API operations a user can perform. For example, users can perform job functions only for jobs started by them. An admin can perform job functions for any job present on the server.

## Future Enhancements

The prototype job worker service will have future enhancements to augment the service and make it production ready.

* While jobs currently run in-memory, job logs will eventually be stored persistently in a database.  
* Bearer token authentication will use an extra layer to process client credentials and validate before sending access tokens to clients.  
* Secrets will be auto-generated and stored securely on both client-side and server-side.  
* The server will use a TLS certificate issued by a trusted Certificate Authority.  
* The authorization scheme allows users to operate on jobs started by them, while admins can operate on any jobs in the system. Future considerations should include an allow-list of programs on the server machine to avoid execution of sensitive programs.  
* For scaling considerations, jobs should be distributed across multiple machines to protect against overload. Resource limits may also be employed per job.  
* Currently running jobs may need to persist beyond server restarts for a robust service.
