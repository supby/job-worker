package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/supby/job-worker/api"
	"github.com/supby/job-worker/workerlib"
	"github.com/supby/job-worker/workerlib/job"
)

func main() {

	err := api.StartServer(api.Configuration{
		Endpoint: "localhost:5001",
	})
	if err != nil {
		log.Fatalf("fail to start server, %v", err)
	}

	// shouldReturn := startStreamin()
	// if shouldReturn {
	// 	return
	// }
}

func startStreamin() bool {
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
			return true
		case d := <-outchan:
			log.Println(string(d))
		}
	}
	return false
}
