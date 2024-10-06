package joblogger

import (
	"context"
	"sync"
)

// Simple In-Memory Job logger for testing purposes
type JobLogger interface {
	Write(p []byte) (n int, err error)
	GetStream(ctx context.Context) chan []byte
}

func New() JobLogger {
	return &jobLogger{
		buf: make([][]byte, 0),
	}
}

type jobLogger struct {
	buf          [][]byte
	bufMtx       sync.Mutex
	listeners    []chan int
	listenersMtx sync.Mutex
}

func (jl *jobLogger) subscribe(ch chan int) func() {
	jl.listenersMtx.Lock()
	defer jl.listenersMtx.Unlock()

	jl.listeners = append(jl.listeners, ch)
	chIndex := len(jl.listeners) - 1

	return func() {
		jl.unsubscribe(chIndex)
	}
}

func (jl *jobLogger) unsubscribe(index int) {
	jl.listenersMtx.Lock()
	defer jl.listenersMtx.Unlock()

	jl.listeners[index] = jl.listeners[len(jl.listeners)-1]
	jl.listeners = jl.listeners[:len(jl.listeners)-1]
}

func (jl *jobLogger) sendWriteSig() {
	jl.listenersMtx.Lock()
	defer jl.listenersMtx.Unlock()

	for _, lis := range jl.listeners {
		lis <- 1
	}
}

func (jl *jobLogger) Write(p []byte) (n int, err error) {
	jl.bufMtx.Lock()
	jl.buf = append(jl.buf, p)
	jl.bufMtx.Unlock()

	jl.sendWriteSig()

	return len(p), nil
}

func (jl *jobLogger) GetStream(ctx context.Context) chan []byte {
	outchan := make(chan []byte)
	writesig := make(chan int)
	unsubscribe := jl.subscribe(writesig)

	go func() {
		defer unsubscribe()
		defer close(outchan)
		defer close(writesig)

		nextStartIndex := jl.flushBuf(0, outchan)
		for {
			select {
			case <-ctx.Done():
				return
			case <-writesig:
				nextStartIndex = jl.flushBuf(nextStartIndex, outchan)
			default:
				continue
			}
		}
	}()
	return outchan
}

func (jl *jobLogger) flushBuf(startIndex int, outchan chan []byte) int {
	jl.bufMtx.Lock()
	nextStartIndex := len(jl.buf)
	for _, v := range jl.buf[startIndex:] {
		outchan <- v
	}
	jl.bufMtx.Unlock()

	return nextStartIndex
}
