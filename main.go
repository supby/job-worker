package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/supby/job-worker/workerlib"
	"github.com/supby/job-worker/workerlib/job"
)

func main() {
	w := workerlib.New()

	// Example of streaming
	jobID, _ := w.Start(job.Command{
		Name: "cat",
		Args: []string{"/dev/random"},
	})

	outchan, _ := w.Stream(jobID)

	sigCh := make(chan os.Signal, 1)
	defer close(sigCh)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)

	for {
		select {
		case <-sigCh:
			log.Println("Exiting application...")
			return
		case d := <-outchan:
			log.Println(string(d))
		}
	}
}
