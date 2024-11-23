package workerlib

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/supby/job-worker/internal/workerlib/job"
)

// ErrJobNotFound is returned when a job with the given ID is not found
var ErrJobNotFound = errors.New("job not found")

// Worker interface responsible for managing jobs
type Worker interface {
	Start(ctx context.Context, command job.Command) (job.JobID, error)
	Stop(ctx context.Context, jobID job.JobID) error
	QueryStatus(ctx context.Context, jobID job.JobID) (*job.Status, error)
	GetStream(ctx context.Context, jobID job.JobID) (<-chan []byte, error)
	Cleanup(ctx context.Context) error
}

type worker struct {
	jobs sync.Map
}

// New creates a new Worker instance
func New() Worker {
	return &worker{}
}

func (w *worker) Start(ctx context.Context, command job.Command) (job.JobID, error) {
	select {
	case <-ctx.Done():
		return job.Nil, ctx.Err()
	default:
		j, err := job.StartNew(command)
		if err != nil {
			return job.Nil, fmt.Errorf("[worker] failed to start job: %w", err)
		}

		jobID := j.GetID()
		w.jobs.Store(jobID, j)

		log.Printf("[worker] Job started: %x", jobID)
		return jobID, nil
	}
}

func (w *worker) Stop(ctx context.Context, jobID job.JobID) error {
	j, err := w.getJob(jobID)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err = j.Stop()
		if err != nil {
			return fmt.Errorf("[worker] failed to stop job %v: %w", jobID, err)
		}
		log.Printf("[worker] Job stopped: %x", jobID)
		return nil
	}
}

func (w *worker) getJob(jobID job.JobID) (job.Job, error) {
	if j, ok := w.jobs.Load(jobID); ok {
		return j.(job.Job), nil
	}
	return nil, ErrJobNotFound
}

func (w *worker) QueryStatus(ctx context.Context, jobID job.JobID) (*job.Status, error) {
	j, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return j.GetStatus(), nil
	}
}

func (w *worker) GetStream(ctx context.Context, jobID job.JobID) (<-chan []byte, error) {
	j, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}

	return j.GetStream(ctx), nil
}

func (w *worker) Cleanup(ctx context.Context) error {
	var err error
	w.jobs.Range(func(key, value interface{}) bool {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return false
		default:
			j := value.(job.Job)
			if cleanupErr := j.Cleanup(ctx); cleanupErr != nil {
				log.Printf("[worker] error cleaning up job %v: %v", key, cleanupErr)
			}
			w.jobs.Delete(key)
			return true
		}
	})
	return err
}
