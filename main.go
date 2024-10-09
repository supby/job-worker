package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/supby/job-worker/internal/workerlib"
	"github.com/supby/job-worker/internal/workerlib/job"
)

func main() {

	w := workerlib.New()

	jobID, _ := w.Start(job.Command{
		Name:      "bash",
		Arguments: []string{"-c", "while true; do date; sleep 1; done"},
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
