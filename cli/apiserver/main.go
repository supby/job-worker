package main

import (
	"log"

	"github.com/supby/job-worker/api"
)

func main() {
	// TODO: move to config
	err := api.StartServer(api.Configuration{
		Endpoint:              "localhost:5001",
		CAFile:                "./cert/rootCA.pem",
		ServerCertificateFile: "./cert/server.crt",
		ServerKeyFile:         "./cert/server.key",
	})
	if err != nil {
		log.Fatalf("fail to start server, %v", err)
	}
}
