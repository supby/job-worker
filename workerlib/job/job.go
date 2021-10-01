package job

import (
	"log"
	"os/exec"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/supby/job-worker/workerlib/joblogger"
)

type Job interface {
	GetID() JobID
	Stop() error
	GetStatus() *Status
	GetStream() chan []byte
}

type job struct {
	id     JobID
	cmd    *exec.Cmd
	status *Status
	logger joblogger.JobLogger
	mtx    sync.Mutex
}

func StartNew(command Command) (Job, error) {
	jobID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	job := job{
		id:     JobID(jobID),
		status: &Status{},
		logger: joblogger.New(),
	}

	cmd := exec.Command(command.Name, command.Args...)
	cmd.Stdout = job.logger
	cmd.Stderr = job.logger

	job.cmd = cmd
	job.status.StatusCode = STARTED

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	job.status.StatusCode = RUNNING

	go job.updateJobStatus()

	return &job, nil
}

func (j *job) GetID() JobID {
	return j.id
}

func (j *job) updateJobStatus() {
	if err := j.cmd.Wait(); err != nil {
		log.Printf("Command execution failed, %v\n", err)
	}

	j.status.ExitCode = j.cmd.ProcessState.ExitCode()
	j.status.Exited = j.cmd.ProcessState.Exited()
	if j.status.StatusCode != STOPPED {
		j.status.StatusCode = EXITED
	}
}

func (j *job) Stop() error {
	j.mtx.Lock()
	defer j.mtx.Unlock()

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
