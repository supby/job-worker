package job

import (
	"log"
	"os/exec"
	"syscall"

	"github.com/google/uuid"
	"github.com/supby/job-worker/workerlib/joblogger"
)

type Job interface {
	GetID() JobID
	Start(cmd Command) error
	Stop() error
	GetStatus() *Status
	GetStream() chan []byte
}

type job struct {
	id     JobID
	cmd    *exec.Cmd
	status *Status
	logger joblogger.JobLogger
}

func New() (Job, error) {
	jobID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &job{
		id:     JobID(jobID),
		status: &Status{},
		logger: joblogger.New(),
	}, nil
}

func (j *job) GetID() JobID {
	return j.id
}

func (j *job) Start(command Command) error {
	cmd := exec.Command(command.Name, command.Args...)
	cmd.Stdout = j.logger
	cmd.Stderr = j.logger

	j.cmd = cmd

	if err := cmd.Start(); err != nil {
		return err
	}

	j.status.StatusCode = STARTED

	go j.updateJobStatus()

	return nil
}

func (j *job) updateJobStatus() {
	if err := j.cmd.Wait(); err != nil {
		log.Printf("Command execution failed, %v\n", err)
	}

	j.status.ExitCode = j.cmd.ProcessState.ExitCode()
	j.status.Exited = j.cmd.ProcessState.Exited()

	if j.status.Exited {
		j.status.StatusCode = EXITED
	} else {
		j.status.StatusCode = RUNNING
	}
}

func (j *job) Stop() error {
	if !j.status.Exited {
		j.status.StatusCode = STOPPED
		return j.cmd.Process.Signal(syscall.SIGKILL)
	}
	return nil
}

func (j *job) GetStatus() *Status {
	return j.status
}

func (j *job) GetStream() chan []byte {
	return j.logger.GetStream()
}
