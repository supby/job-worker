package job

const (
	UNKNOWN = 0
	RUNNING = 1
	EXITED  = 2
	STOPPED = 3
	STARTED = 4
)

type JobID [16]byte

var Nil JobID // empty UUID, all zeros

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
}
