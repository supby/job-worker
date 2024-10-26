package job

import (
	"context"
	"log"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/google/uuid"
	"github.com/supby/job-worker/internal/workerlib/joblogger"
)

// Job interface encapsulates logic for one job.
type Job interface {
	GetID() JobID
	Stop() error
	GetStatus() *Status
	GetStream(ctx context.Context) <-chan []byte
	Cleanup(ctx context.Context) error
}

type job struct {
	id     JobID
	cmd    *exec.Cmd
	status atomic.Value
	logger joblogger.JobLogger
	mtx    sync.Mutex
}

func StartNew(command Command) (Job, error) {
	jobID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	logger, err := joblogger.New()
	if err != nil {
		return nil, err
	}

	j := &job{
		id:     JobID(jobID),
		logger: logger,
	}

	status := &Status{
		CommandName: command.Name,
		Arguments:   command.Arguments,
		StatusCode:  STARTED,
	}
	j.status.Store(status)

	cmd := exec.Command(command.Name, command.Arguments...)
	cmd.Stdout = j.logger
	cmd.Stderr = j.logger

	j.cmd = cmd

	if err := cmd.Start(); err != nil {
		j.updateStatus(func(s *Status) {
			s.StatusCode = ERROR
			s.Error = err.Error()
		})
		return nil, err
	}

	j.updateStatus(func(s *Status) {
		s.StatusCode = RUNNING
	})

	go j.updateJobStatus()

	return j, nil
}

func (j *job) GetID() JobID {
	return j.id
}

func (j *job) updateJobStatus() {
	err := j.cmd.Wait()
	j.updateStatus(func(s *Status) {
		s.ExitCode = j.cmd.ProcessState.ExitCode()
		s.Exited = j.cmd.ProcessState.Exited()
		if s.StatusCode != STOPPED {
			s.StatusCode = EXITED
		}
		if err != nil {
			log.Printf("Command execution failed: %v", err)
			s.Error = err.Error()
		}
	})
}

func (j *job) Stop() error {
	j.mtx.Lock()
	defer j.mtx.Unlock()

	if j.cmd.ProcessState == nil || !j.cmd.ProcessState.Exited() {
		j.updateStatus(func(s *Status) {
			s.StatusCode = STOPPED
		})
		return j.cmd.Process.Signal(syscall.SIGKILL)
	}
	return nil
}

func (j *job) GetStatus() *Status {
	return j.status.Load().(*Status)
}

func (j *job) GetStream(ctx context.Context) <-chan []byte {
	return j.logger.GetStream(ctx)
}

func (j *job) updateStatus(updateFn func(*Status)) {
	for {
		oldStatus := j.status.Load().(*Status)
		newStatus := *oldStatus // Create a copy
		updateFn(&newStatus)
		if j.status.CompareAndSwap(oldStatus, &newStatus) {
			break
		}
	}
}

func (j *job) Cleanup(ctx context.Context) error {
	return nil
}
