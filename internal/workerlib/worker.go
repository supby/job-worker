package workerlib

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/supby/job-worker/workerlib/job"
)

// Worker interface responsible for jobs managing
type Worker interface {
	Start(command job.Command) (job.JobID, error)
	Stop(jobID job.JobID) error
	QueryStatus(jobID job.JobID) (*job.Status, error)
	GetStream(ctx context.Context, jobID job.JobID) (<-chan []byte, error)
}

type worker struct {
	jobs sync.Map
}

func New() Worker {
	return &worker{}
}

func (w *worker) Start(command job.Command) (job.JobID, error) {
	j, err := job.StartNew(command)
	if err != nil {
		log.Printf("Job starting failed: %v", err)
		return job.Nil, fmt.Errorf("failed to start job: %w", err)
	}

	jobID := j.GetID()
	w.jobs.Store(jobID, j)

	return jobID, nil
}

func (w *worker) Stop(jobID job.JobID) error {
	job, err := w.getJob(jobID)
	if err != nil {
		return err
	}

	return job.Stop()
}

func (w *worker) getJob(jobID job.JobID) (job.Job, error) {
	if ret, ok := w.jobs.Load(jobID); ok {
		return ret.(job.Job), nil
	}
	return nil, fmt.Errorf("job %v not found", jobID)
}

func (w *worker) QueryStatus(jobID job.JobID) (*job.Status, error) {
	job, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}
	return job.GetStatus(), nil
}

func (w *worker) GetStream(ctx context.Context, jobID job.JobID) (<-chan []byte, error) {
	job, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}
	return job.GetStream(ctx), nil
}
