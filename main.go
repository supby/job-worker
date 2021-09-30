package main

import (
	"log"

	"github.com/supby/job-worker/workerlib"
	"github.com/supby/job-worker/workerlib/job"
)

func main() {
	w1 := workerlib.New()
	jobID, _ := w1.Start(job.Command{
		Name: "ls",
		Args: []string{"-l"},
	})

	outchan, _ := w1.Stream(jobID)

	d := <-outchan

	log.Println(string(d))
}
