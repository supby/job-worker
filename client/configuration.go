package client

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	ServerEndpoint        string
	CAFile                string
	ClientCertificateFile string
	ClientKeyFile         string
}

func LoadConfigFromYaml(filename string) Configuration {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		log.Fatalf("Configuration file '%v' does not exist.", filename)
	}

	data, _ := os.ReadFile(filename)

	cfg := Configuration{}

	err = yaml.Unmarshal([]byte(data), &cfg)
	if err != nil {
		log.Fatalf("error loading YAML config: %v", err)
	}

	if cfg.ServerEndpoint == "" {
		log.Fatal("Server Endpoint is empty")
	}

	return cfg
}
