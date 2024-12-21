package joblogger

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestGetStreamExistingJob(t *testing.T) {
	jobID, _ := uuid.NewRandom()
	jl, err := New(jobID)
	assert.NoError(t, err)
	defer jl.Close()

	// Write some logs
	_, err = jl.Write([]byte("log line 1\n"))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outchan, err := jl.GetStream(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, outchan)

	select {
	case log := <-outchan:
		assert.Contains(t, string(log), "log line 1")
	case <-time.After(time.Second):
		t.Fatal("expected log line, but got timeout")
	}
}

func TestGetStreamNoLogs(t *testing.T) {
	jobID, _ := uuid.NewRandom()
	jl, err := New(jobID)
	assert.NoError(t, err)
	defer jl.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outchan, err := jl.GetStream(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, outchan)

	select {
	case <-outchan:
		assert.Fail(t, "it should be any logs")
	case <-time.After(time.Second):
		t.Log("expected timeout as there is no logs")
		return
	}

	assert.Fail(t, "should return on timeou")
}

func TestGetStreamMultipleListeners(t *testing.T) {
	jobID, _ := uuid.NewRandom()
	jl, err := New(jobID)
	assert.NoError(t, err)
	defer jl.Close()

	// Write some logs
	_, err = jl.Write([]byte("log line 1\n"))
	assert.NoError(t, err)

	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()
	outchan1, err := jl.GetStream(ctx1)
	assert.NoError(t, err)
	assert.NotNil(t, outchan1)

	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	outchan2, err := jl.GetStream(ctx2)
	assert.NoError(t, err)
	assert.NotNil(t, outchan2)

	select {
	case log := <-outchan1:
		assert.Contains(t, string(log), "log line 1")
	case <-time.After(time.Second):
		t.Fatal("expected log line, but got timeout")
	}

	select {
	case log := <-outchan2:
		assert.Contains(t, string(log), "log line 1")
	case <-time.After(time.Second):
		t.Fatal("expected log line, but got timeout")
	}
}

func TestGetStreamAndCloseLogger(t *testing.T) {
	jobID, _ := uuid.NewRandom()
	jl, err := New(jobID)
	assert.NoError(t, err)
	defer jl.Close()

	// Write some logs
	_, err = jl.Write([]byte("log line 1\n"))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outchan, err := jl.GetStream(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, outchan)

	select {
	case log := <-outchan:
		assert.Contains(t, string(log), "log line 1")
	case <-time.After(time.Second):
		t.Fatal("expected log line, but got timeout")
	}

	// Close the job logger
	err = jl.Close()
	assert.NoError(t, err)

	select {
	case <-outchan:
		assert.Fail(t, "should not be any logs")
	case <-time.After(time.Second):
		t.Log("expected timeout as logger is closed")
		return
	}

	assert.Fail(t, "should return on timout")
}

func TestGetStreamAfterLoggerClosed(t *testing.T) {
	jobID, _ := uuid.NewRandom()
	jl, err := New(jobID)
	assert.NoError(t, err)
	defer jl.Close()

	// Write some logs
	_, err = jl.Write([]byte("log line 1\n"))
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Close the job logger
	err = jl.Close()
	assert.NoError(t, err)

	_, err = jl.GetStream(ctx)
	assert.Error(t, err)
}
