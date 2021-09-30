package main

import (
	"log"

	"github.com/supby/job-worker/workerlib"
	"github.com/supby/job-worker/workerlib/job"
)

func main() {
	w := workerlib.New()

	// Example of streaming
	jobID, _ := w.Start(job.Command{
		Name: "ls",
		Args: []string{"-l"},
	})

	outchan, _ := w.Stream(jobID)

	d := <-outchan

	log.Println(string(d))
}
