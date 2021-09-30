package joblogger

type JobLogger interface {
	Write(p []byte) (n int, err error)
	GetStream() chan []byte
}

func New() JobLogger {
	// TODO: channel is buffered as it is in memory store.
	return &jobLogger{
		outchan: make(chan []byte, 128),
	}
}

type jobLogger struct {
	outchan chan []byte
}

func (jl *jobLogger) Write(p []byte) (n int, err error) {
	jl.outchan <- p

	return len(p), nil
}

func (jl *jobLogger) GetStream() chan []byte {
	return jl.outchan
}

func (jl *jobLogger) dispose() {
	close(jl.outchan)
}