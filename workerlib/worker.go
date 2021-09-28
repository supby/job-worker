package workerlib

type IWorker interface {
	Start(command Command) (jobID string, err error)
	Stop(jobID []byte) (err error)
	Query(jobID []byte) (status Status, err error)
	Stream(jobID []byte) (logchan chan string, err error)
}

type worker struct {
	jobs map[string]*Job
}
