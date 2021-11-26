package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/supby/job-worker/cli/client/argsparser"
	"github.com/supby/job-worker/client"
	"github.com/supby/job-worker/generated/proto"
)

func main() {
	// TODO: move it to config
	cfg := client.Configuration{
		ServerEndpoint:        "localhost:5001",
		CAFile:                "./cert/rootCA.pem",
		ClientCertificateFile: "./cert/server.crt",
		ClientKeyFile:         "./cert/server.key",
	}

	parameters, err := argsparser.GetParams(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing CLI parameters: %v", err)
	}

	wsclient, err := client.NewWorkerClient(cfg)
	if err != nil {
		log.Fatalf("Error creating client %v", err)
	}

	pctx := context.Background()
	ctx, cancel := context.WithTimeout(pctx, time.Duration(1000)*time.Millisecond)
	defer cancel()

	switch parameters.CLICommand {
	case argsparser.START_COMMAND:
		handleStartCommand(ctx, wsclient, parameters)
	case argsparser.STOP_COMMAND:
		handleStopCommand(ctx, wsclient, parameters)
	case argsparser.QUERY_COMMAND:
		handleQueryCommand(ctx, wsclient, parameters)
	case argsparser.STREAM_COMMAND:
		handleStreamCommand(pctx, wsclient, parameters)
	}
}

func handleQueryCommand(ctx context.Context, wsclient proto.WorkerServiceClient, parameters *argsparser.Parameters) {
	jobID, _ := hex.DecodeString(parameters.JobID)
	resp, err := wsclient.QueryStatus(ctx, &proto.QueryStatusRequest{
		JobID: jobID,
	})
	if err != nil {
		log.Fatalf("Error QueryStatus command %v", err)
	}

	log.Printf("QueryStatus Resp: %v", resp)
}

func handleStreamCommand(ctx context.Context, wsclient proto.WorkerServiceClient, parameters *argsparser.Parameters) {
	jobID, _ := hex.DecodeString(parameters.JobID)

	ctx, cancel := context.WithCancel(ctx)
	resp, err := wsclient.GetOutput(ctx, &proto.GetOutputRequest{
		JobID: jobID,
	})
	if err != nil {
		log.Fatalf("Error stream: %v", err)
	}

	go func() {
		log.Println("Job output stream:")
		for {
			out, err := resp.Recv()
			if err != nil {
				log.Printf("Error stream: %v", err)
			}
			fmt.Print(string(out.Output))
		}
	}()

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	defer func() {
		cancel()
		signal.Stop(sigchan)
	}()
	<-sigchan
}

func handleStopCommand(ctx context.Context, wsclient proto.WorkerServiceClient, parameters *argsparser.Parameters) {
	jobID, _ := hex.DecodeString(parameters.JobID)
	resp, err := wsclient.Stop(ctx, &proto.StopRequest{
		JobID: jobID,
	})
	if err != nil {
		log.Fatalf("Error Stop command %v", err)
	}

	log.Printf("Stop Resp: %v", resp)
}

func handleStartCommand(ctx context.Context, wsclient proto.WorkerServiceClient, parameters *argsparser.Parameters) {
	resp, err := wsclient.Start(ctx, &proto.StartRequest{
		CommandName: parameters.CommandName,
		Arguments:   parameters.Arguments,
	})
	if err != nil {
		log.Fatalf("Error start command %v", err)
	}

	log.Printf("Started JobID: %x\n", resp.GetJobID())
}
