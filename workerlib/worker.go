package workerlib

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/supby/job-worker/workerlib/job"
)

// Worker interface resposible for jobs managing
type Worker interface {
	Start(command job.Command) (job.JobID, error)
	Stop(jobID job.JobID) error
	QueryStatus(jobID job.JobID) (*job.Status, error)
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
	j, err := job.StartNew(command)
	if err != nil {
		log.Printf("Job staring failed, %v\n", err)
		return job.Nil, err
	}

	jobID := j.GetID()

	w.mtx.Lock()
	w.jobs[jobID] = j
	w.mtx.Unlock()

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
	w.mtx.Lock()
	job, found := w.jobs[jobID]
	w.mtx.Unlock()

	if !found {
		msg := fmt.Sprintf("Job %v is not found", jobID)
		log.Println(msg)
		return nil, errors.New(msg)
	}
	return job, nil
}

func (w *worker) QueryStatus(jobID job.JobID) (*job.Status, error) {
	job, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}
	return job.GetStatus(), nil
}

func (w *worker) Stream(jobID job.JobID) (chan []byte, error) {
	job, err := w.getJob(jobID)
	if err != nil {
		return nil, err
	}
	return job.GetStream(), nil
}
