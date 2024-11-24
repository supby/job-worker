package job

import "github.com/google/uuid"

const (
	UNKNOWN = 0
	RUNNING = 1
	EXITED  = 2
	STOPPED = 3
	STARTED = 4
	ERROR   = 5
)

var NilJobId uuid.UUID // empty UUID, all zeros

type Command struct {
	Name      string
	Arguments []string
}

type Status struct {
	ExitCode    int
	Exited      bool
	StatusCode  byte
	CommandName string
	Arguments   []string
	Error       string
}
