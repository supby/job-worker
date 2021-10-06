package job

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/supby/job-worker/workerlib/joblogger"
)

// Job interface encapsulate logic for one job.
type Job interface {
	GetID() JobID
	Stop() error
	GetStatus() *Status
	GetStream(ctx context.Context) chan []byte
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
		// TODO: Error here ccould be reflected on Job status with new ERROR status.
		return nil, err
	}

	// TODO: here is a bit artificial case when job STARTED initially and RUNNING after cmd.Start
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

	j.mtx.Lock()
	defer j.mtx.Unlock()

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

func (j *job) GetStream(ctx context.Context) chan []byte {
	return j.logger.GetStream(ctx)
}
