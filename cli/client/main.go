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
	ctx, _ := context.WithTimeout(pctx, time.Duration(1000)*time.Millisecond)

	switch parameters.CLICommand {
	case argsparser.START_COMMAND:
		resp, err := wsclient.Start(ctx, &proto.StartRequest{
			CommandName: parameters.CommandName,
			Arguments:   parameters.Arguments,
		})
		if err != nil {
			log.Fatalf("Error start command %v", err)
		}

		log.Printf("Started JobID: %x\n", resp.GetJobID())
		break
	case argsparser.STOP_COMMAND:

		resp, err := wsclient.Stop(ctx, &proto.StopRequest{})
		if err != nil {
			log.Fatalf("Error Stop command %v", err)
		}

		log.Printf("Stop Resp: %v", resp)
		break
	case argsparser.QUERY_COMMAND:
		//wsclient.QueryStatus()
		break
	case argsparser.STREAM_COMMAND:
		jobID, err := hex.DecodeString(parameters.JobID)
		if err != nil {
			log.Fatalf("Error parse JobID: %v", err)
		}
		ctx, cancel := context.WithCancel(pctx)
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
					return
				}
				fmt.Print(out.String())
			}
		}()

		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		defer func() {
			cancel()
			signal.Stop(sigchan)
		}()
		<-sigchan

		break
	}
}
