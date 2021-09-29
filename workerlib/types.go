package workerlib

import (
	"os/exec"

	"github.com/supby/job-worker/workerlib/joblogger"
)

type Command struct {
	Name string
	Args []string
}

type Status struct {
	ExitCode int
	Exited   bool
}

type Job struct {
	ID     [16]byte
	Cmd    *exec.Cmd
	Status *Status
	Logger joblogger.JobLogger
}
