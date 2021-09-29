package workerlib

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"syscall"

	"github.com/google/uuid"
	"github.com/supby/job-worker/workerlib/joblogger"
)

type Worker interface {
	Start(command Command) ([16]byte, error)
	Stop(jobID [16]byte) error
	Query(jobID [16]byte) (Status, error)
	Stream(jobID [16]byte) (chan []byte, error)
}

func New() Worker {
	return &worker{
		jobs: make(map[[16]byte]*Job),
	}
}

type worker struct {
	jobs map[[16]byte]*Job
	mtx  sync.Mutex
}

func (w *worker) Start(command Command) ([16]byte, error) {
	cmd := exec.Command(command.Name, command.Args...)
	jobID, err := uuid.NewRandom()
	if err != nil {
		log.Printf("JobID creation failed, %v\n", err)
		return jobID, err
	}

	jobLogger := joblogger.New()

	cmd.Stdout = jobLogger
	cmd.Stderr = jobLogger
	if err = cmd.Start(); err != nil {
		log.Printf("Command starting failed, %v\n", err)
		return jobID, err
	}

	job := Job{ID: jobID, Cmd: cmd, Status: &Status{}, Logger: jobLogger}
	w.mtx.Lock()
	w.jobs[jobID] = &job
	w.mtx.Unlock()

	go w.updateJobStatus(job)

	return jobID, nil
}

func (w *worker) updateJobStatus(job Job) {
	if err := job.Cmd.Wait(); err != nil {
		log.Printf("Command execution failed, %v\n", err)
	}

	status := Status{
		ExitCode: job.Cmd.ProcessState.ExitCode(),
		Exited:   job.Cmd.ProcessState.Exited(),
	}
	w.mtx.Lock()
	job.Status = &status
	w.mtx.Unlock()
}

func (w *worker) Stop(jobID [16]byte) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	job, err := w.getJob(jobID)
	if err != nil {
		return err
	}

	if !job.Status.Exited {
		return job.Cmd.Process.Signal(syscall.SIGKILL)
	}
	return nil
}

func (w *worker) getJob(jobID [16]byte) (*Job, error) {
	job, found := w.jobs[jobID]
	if !found {
		msg := fmt.Sprintf("Job %v is not found", jobID)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	return job, nil
}

func (w *worker) Query(jobID [16]byte) (Status, error) {
	w.mtx.Lock()
	job, err := w.getJob(jobID)
	w.mtx.Unlock()
	if err != nil {
		return Status{}, err
	}
	return *job.Status, nil
}

func (w *worker) Stream(jobID [16]byte) (chan []byte, error) {
	w.mtx.Lock()
	job, err := w.getJob(jobID)
	w.mtx.Unlock()
	if err != nil {
		return nil, err
	}
	return job.Logger.GetStream(), nil
}
