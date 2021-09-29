package main

import (
	"log"

	"github.com/supby/job-worker/workerlib"
)

func main() {
	w1 := workerlib.New()
	jobID, _ := w1.Start(workerlib.Command{
		Name: "ls",
		Args: []string{"-l"},
	})

	outchan, _ := w1.Stream(jobID)

	d := <-outchan

	log.Println(d)
}
