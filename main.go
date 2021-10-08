package main

import (
	"context"
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

	// jobID, _ := w.Start(job.Command{
	// 	Name: "cat",
	// 	Args: []string{"/dev/random"},
	// })

	// jobID, _ := w.Start(job.Command{
	// 	Name: "seq",
	// 	Args: []string{"1000000"},
	// })

	jobID, _ := w.Start(job.Command{
		Name: "bash",
		Args: []string{"-c", "while true; do date; sleep 1; done"},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outchan, _ := w.GetStream(ctx, jobID)

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
