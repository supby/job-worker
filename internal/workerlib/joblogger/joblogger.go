package joblogger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"github.com/google/uuid"
)

// JobLogger is an interface for a job logger that uses a temporary file for storage
type JobLogger interface {
	Write(p []byte) (n int, err error)
	GetStream(ctx context.Context) <-chan []byte
	Close() error
}

type listener struct {
	offset int64
	notify chan struct{}
}

type jobLogger struct {
	jobId     uuid.UUID
	file      *os.File
	mu        sync.Mutex
	listeners map[*listener]struct{}
}

func New(jobId uuid.UUID) (JobLogger, error) {
	file, err := os.CreateTemp("", fmt.Sprintf("joblog-%x-*.txt", jobId))
	if err != nil {
		return nil, err
	}

	return &jobLogger{
		jobId:     jobId,
		file:      file,
		listeners: make(map[*listener]struct{}),
	}, nil
}

func (jl *jobLogger) Write(p []byte) (n int, err error) {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	n, err = jl.file.Write(p)
	if err != nil {
		return n, err
	}

	jl.notifyListeners()

	return n, nil
}

func (jl *jobLogger) notifyListeners() {
	for l := range jl.listeners {
		select {
		case l.notify <- struct{}{}:
		default:
			// If the channel is full, we skip the notification
			// This prevents blocking if a listener is slow
		}
	}
}

func (jl *jobLogger) GetStream(ctx context.Context) <-chan []byte {
	outchan := make(chan []byte, 100) // Buffered channel to reduce blocking
	l := &listener{
		offset: 0,
		notify: make(chan struct{}, 1),
	}

	jl.mu.Lock()
	jl.listeners[l] = struct{}{}
	jl.mu.Unlock()

	go func() {
		defer func() {
			jl.mu.Lock()
			delete(jl.listeners, l)
			jl.mu.Unlock()
			close(outchan)
			close(l.notify)
		}()

		// in case log file already contains something
		if err := jl.flushToChannel(l, outchan); err != nil {
			log.Printf("joblogger] job logs flushing failed, jobId: %x, error: %v", jl.jobId, err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case <-l.notify:
				if err := jl.flushToChannel(l, outchan); err != nil {
					log.Printf("joblogger] job logs flushing failed, jobId: %x, error: %v", jl.jobId, err)
					return
				}
			}
		}
	}()

	return outchan
}

func (jl *jobLogger) flushToChannel(l *listener, outchan chan<- []byte) error {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	_, err := jl.file.Seek(l.offset, io.SeekStart)
	if err != nil {
		return err
	}

	buffer := make([]byte, 4096)
	for {
		n, err := jl.file.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		select {
		case outchan <- buffer[:n]:
			l.offset += int64(n)
		default:
			// If the channel is full, we return early
			// This prevents blocking if the consumer is slow
			return nil
		}
	}

	return nil
}

func (jl *jobLogger) Close() error {
	jl.mu.Lock()
	defer jl.mu.Unlock()

	if err := jl.file.Close(); err != nil {
		return err
	}

	return os.Remove(jl.file.Name())
}
