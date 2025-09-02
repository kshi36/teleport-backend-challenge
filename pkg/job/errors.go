package job

import "errors"

var ErrJobNotFound = errors.New("job not found")
var ErrJobProcessNull = errors.New("job has not started the process")
var ErrJobNotCompleted = errors.New("job has not completed")
