package main

import (
	"log"

	"github.com/supby/job-worker/api"
)

func main() {
	err := api.StartServer(api.Configuration{
		Endpoint: "localhost:5001",
	})
	if err != nil {
		log.Fatalf("fail to start server, %v", err)
	}
}
