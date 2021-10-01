package workerlib

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/supby/job-worker/workerlib/job"
)

func TestStartExistingCommand(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "ls", Args: []string{"-a"}})

	assert.NotEmpty(t, jobID)
	assert.Nil(t, err)
}

func TestStartNotExistingCommand(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "blablabla17"})

	assert.NotEmpty(t, jobID)
	assert.NotNil(t, err)
}

func TestStopNotExistingJob(t *testing.T) {
	randomJobID, _ := uuid.NewRandom()
	w := New()
	err := w.Stop(job.JobID(randomJobID))

	assert.NotNil(t, err)
}

func TestStopExistingJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "sleep", Args: []string{"2"}})
	assert.Nil(t, err)

	err = w.Stop(jobID)
	assert.Nil(t, err)
}

func TestStopExitedJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "sleep", Args: []string{"1"}})
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)

	err = w.Stop(jobID)
	assert.Nil(t, err)
}

func TestQueryRunningJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "sleep", Args: []string{"1"}})
	assert.Nil(t, err)

	status, err := w.Query(jobID)
	assert.Nil(t, err)
	assert.False(t, status.Exited)
	assert.True(t, status.StatusCode == job.RUNNING)
}

func TestQueryExitedJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "sleep", Args: []string{"1"}})
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)

	status, err := w.Query(jobID)
	assert.Nil(t, err)
	assert.True(t, status.Exited)
	assert.True(t, status.StatusCode == job.EXITED)
}

func TestQueryStoppedJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "sleep", Args: []string{"1"}})
	assert.Nil(t, err)

	err = w.Stop(jobID)
	assert.Nil(t, err)

	time.Sleep(time.Second * 2)

	status, err := w.Query(jobID)
	assert.Nil(t, err)
	assert.False(t, status.Exited)
	assert.True(t, status.StatusCode == job.STOPPED)
}

func TestQueryNotExistingJob(t *testing.T) {
	randomJobID, _ := uuid.NewRandom()
	w := New()
	status, err := w.Query(job.JobID(randomJobID))

	assert.NotNil(t, err)
	assert.Nil(t, status)
}

func TestStreamExistingJob(t *testing.T) {
	w := New()
	jobID, err := w.Start(job.Command{Name: "bash", Args: []string{"-c", "while true; do date; sleep 1; done"}})
	assert.Nil(t, err)

	outchan, err := w.Stream(jobID)
	assert.Nil(t, err)
	assert.NotNil(t, <-outchan)

	err = w.Stop(jobID)
	assert.Nil(t, err)
}

func TestStreamNotExistingJob(t *testing.T) {
	randomJobID, _ := uuid.NewRandom()
	w := New()
	outchan, err := w.Stream(job.JobID(randomJobID))

	assert.Nil(t, outchan)
	assert.Error(t, err)
}
