package api

import (
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Endpoint              string
	CAFile                string
	ServerCertificateFile string
	ServerKeyFile         string
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

	if cfg.Endpoint == "" {
		log.Println("Endpoint is empty in configuration, using default 127.0.0.1:5001")
		cfg.Endpoint = "127.0.0.1:5001"
	}

	return cfg
}
