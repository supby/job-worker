package main

import (
	"log"
	"os"

	"github.com/supby/job-worker/cli/client/argsparser"
	"github.com/supby/job-worker/client"
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

	switch parameters.CLICommand {
	case argsparser.START_COMMAND:
		//wsclient.Start()
		break
	case argsparser.STOP_COMMAND:
		//wsclient.Stop()
		break
	case argsparser.QUERY_COMMAND:
		//wsclient.QueryStatus()
		break
	case argsparser.STREAM_COMMAND:
		//wsclient.GetOutput()
		break
	}

	log.Println("Client CLI")
}
