package workerlib

import "os/exec"

type Command struct {
	Name string
	Args []string
}

type Status struct {
	ExitCode int
	Exited   bool
}

type Job struct {
	ID     []byte
	Cmd    *exec.Cmd
	Status *Status
}
