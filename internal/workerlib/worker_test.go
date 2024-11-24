package workerlib

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/supby/job-worker/internal/workerlib/job"
)

func TestStartExistingCommand(t *testing.T) {
	testCtx := context.Background()

	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "ls", Arguments: []string{"-a"}})

	assert.NotEmpty(t, jobID)
	assert.NoError(t, err)
}

func TestStartNotExistingCommand(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "blablabla17"})

	assert.NotEmpty(t, jobID)
	assert.Error(t, err)
}

func TestStopNotExistingJob(t *testing.T) {
	testCtx := context.Background()
	randomJobID, _ := uuid.NewRandom()
	w := New()
	err := w.Stop(testCtx, randomJobID)

	assert.Error(t, err)
}

func TestStopExistingJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "sleep", Arguments: []string{"2"}})
	assert.NoError(t, err)

	err = w.Stop(testCtx, jobID)
	assert.NoError(t, err)
}

func TestStopExitedJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "sleep", Arguments: []string{"1"}})
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)

	err = w.Stop(testCtx, jobID)
	assert.NoError(t, err)
}

func TestQueryRunningJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "sleep", Arguments: []string{"1"}})
	assert.NoError(t, err)

	status, err := w.QueryStatus(testCtx, jobID)
	assert.NoError(t, err)
	assert.False(t, status.Exited)
	assert.True(t, status.StatusCode == job.RUNNING)
}

func TestQueryExitedJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "sleep", Arguments: []string{"1"}})
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)

	status, err := w.QueryStatus(testCtx, jobID)
	assert.NoError(t, err)
	assert.True(t, status.Exited)
	assert.True(t, status.StatusCode == job.EXITED)
}

func TestQueryStoppedJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "sleep", Arguments: []string{"1"}})
	assert.NoError(t, err)

	err = w.Stop(testCtx, jobID)
	assert.NoError(t, err)

	time.Sleep(time.Second * 2)

	status, err := w.QueryStatus(testCtx, jobID)
	assert.NoError(t, err)
	assert.False(t, status.Exited)
	assert.True(t, status.StatusCode == job.STOPPED)
}

func TestQueryNotExistingJob(t *testing.T) {
	testCtx := context.Background()
	randomJobID, _ := uuid.NewRandom()
	w := New()
	status, err := w.QueryStatus(testCtx, randomJobID)

	assert.Error(t, err)
	assert.Nil(t, status)
}

func TestStreamExistingJob(t *testing.T) {
	testCtx := context.Background()
	w := New()
	jobID, err := w.Start(testCtx, job.Command{Name: "bash", Arguments: []string{"-c", "while true; do date; sleep 1; done"}})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outchan, err := w.GetStream(ctx, jobID)
	assert.NoError(t, err)
	assert.NotNil(t, <-outchan)

	err = w.Stop(testCtx, jobID)
	assert.NoError(t, err)
}

func TestStreamNotExistingJob(t *testing.T) {
	randomJobID, _ := uuid.NewRandom()
	w := New()
	outchan, err := w.GetStream(context.Background(), randomJobID)

	assert.Nil(t, outchan)
	assert.Error(t, err)
}
