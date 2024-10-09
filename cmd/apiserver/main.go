package main

import (
	"log"

	"github.com/supby/job-worker/internal/api"
)

func main() {
	cfg := api.LoadConfigFromYaml("./server_config.yaml")
	err := api.StartServer(cfg)
	if err != nil {
		log.Fatalf("fail to start server, %v", err)
	}
}
