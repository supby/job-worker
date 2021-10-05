package joblogger

import (
	"context"
	"sync"
)

// Simple In-Memory Job logger for testing purposes
type JobLogger interface {
	Write(p []byte) (n int, err error)
	GetStream(ctx context.Context) chan []byte
	Dispose()
}

func New() JobLogger {
	return &jobLogger{
		buf:      make([][]byte, 0),
		writesig: make(chan int, 1),
	}
}

type jobLogger struct {
	buf      [][]byte
	mtx      sync.Mutex
	writesig chan int
}

func (jl *jobLogger) Write(p []byte) (n int, err error) {
	jl.mtx.Lock()
	jl.buf = append(jl.buf, p)
	jl.mtx.Unlock()

	select {
	case jl.writesig <- 1:
	default:
	}

	return len(p), nil
}

func (jl *jobLogger) GetStream(ctx context.Context) chan []byte {
	outchan := make(chan []byte)
	go func() {
		nextStartIndex := jl.flushBuf(0, outchan)
		for {
			select {
			case <-ctx.Done():
				close(outchan)
				return
			case <-jl.writesig:
				nextStartIndex = jl.flushBuf(nextStartIndex, outchan)
			default:
				continue
			}
		}
	}()
	return outchan
}

func (jl *jobLogger) flushBuf(startIndex int, outchan chan []byte) int {
	jl.mtx.Lock()
	nextStartIndex := len(jl.buf)
	for _, v := range jl.buf[startIndex:] {
		outchan <- v
	}
	jl.mtx.Unlock()

	return nextStartIndex
}

func (jl *jobLogger) Dispose() {
	close(jl.writesig)
}
