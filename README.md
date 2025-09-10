# teleport-backend-challenge
Teleport Backend Challenge - Level 1


## Makefile
`make all` - Build jobserver and jobctl binaries

`make clean` - Clean up binaries

`make test` - Perform go tests with go race detector

## CLI tool
The CLI tool `jobctl` provides an interface to perform HTTPS requests to the API server. This includes job management functions such as start a job, stop a job, get status, and get output.

In addition, the `-u (--user)` flag allows a user to change their user ID for testing purposes (Usage: `-u fakeuser`).

The `jobserver` program starts up the API server to receive HTTPS requests.

### Example Usage
Start the job server

`./jobserver`

Start a job, receive a new ID

`./jobctl start -- /bin/sleep 5`  
`Job started with ID j-12345`

Stop a job by ID

`./jobctl status j-12345`  
`Job status for ID j-12345`  
`Status: Running`  
`Exit code:`

Get status of a job by ID

`./jobctl stop j-12345`  
`Job stopped for ID j-12345`

Get the output of a job

`./jobctl output j-98765`  
`Job output for ID j-98765`  
`stdout:`   
`hello world`  
`stderr:`
