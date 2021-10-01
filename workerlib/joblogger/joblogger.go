package joblogger

// Simple In-Memory Job logger for testing purposes
type JobLogger interface {
	Write(p []byte) (n int, err error)
	GetStream() chan []byte
	Dispose()
}

func New() JobLogger {
	// TODO: channel size is harcoded here as it is POC in-memory logger.
	return &jobLogger{
		outchan: make(chan []byte, 20),
	}
}

type jobLogger struct {
	outchan chan []byte
}

func (jl *jobLogger) Write(p []byte) (n int, err error) {
	// TODO: in case if channel is full it takes out one item to free space. Just for test task purposes.
	if len(jl.outchan) == cap(jl.outchan) {
		<-jl.outchan
	}

	jl.outchan <- p

	return len(p), nil
}

func (jl *jobLogger) GetStream() chan []byte {
	return jl.outchan
}

func (jl *jobLogger) Dispose() {
	close(jl.outchan)
}
