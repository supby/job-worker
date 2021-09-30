package workerlib

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/supby/job-worker/workerlib/job"
)

type Worker interface {
	Start(command job.Command) (job.JobID, error)
	Stop(jobID job.JobID) error
	Query(jobID job.JobID) (job.Status, error)
	Stream(jobID job.JobID) (chan []byte, error)
}

func New() Worker {
	return &worker{
		jobs: make(map[job.JobID]job.Job),
	}
}

type worker struct {
	jobs map[job.JobID]job.Job
	mtx  sync.Mutex
}

func (w *worker) Start(command job.Command) (job.JobID, error) {
	job, err := job.New()
	if err != nil {
		log.Printf("Job creation failed, %v\n", err)
		return job.GetID(), err
	}

	err = job.Start(command)
	if err != nil {
		log.Printf("Job staring failed, %v\n", err)
		return job.GetID(), err
	}

	jobID := job.GetID()

	w.mtx.Lock()
	w.jobs[jobID] = job
	w.mtx.Unlock()

	return jobID, nil
}

func (w *worker) Stop(jobID job.JobID) error {
	w.mtx.Lock()
	defer w.mtx.Unlock()

	job, err := w.getJob(jobID)
	if err != nil {
		return err
	}

	return job.Stop()
}

func (w *worker) getJob(jobID job.JobID) (job.Job, error) {
	job, found := w.jobs[jobID]
	if !found {
		msg := fmt.Sprintf("Job %v is not found", jobID)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	return job, nil
}

func (w *worker) Query(jobID job.JobID) (job.Status, error) {
	w.mtx.Lock()
	job, err := w.getJob(jobID)
	w.mtx.Unlock()
	if err != nil {
		return *job.GetStatus(), err
	}
	return *job.GetStatus(), nil
}

func (w *worker) Stream(jobID job.JobID) (chan []byte, error) {
	w.mtx.Lock()
	job, err := w.getJob(jobID)
	w.mtx.Unlock()
	if err != nil {
		return nil, err
	}
	return job.GetStream(), nil
}
