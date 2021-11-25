package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/supby/job-worker/generated/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func loadTLSCredentials(config Configuration) (credentials.TransportCredentials, error) {
	pemServerCA, err := ioutil.ReadFile(config.CAFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate %v", pemServerCA)
	}

	clientCert, err := tls.LoadX509KeyPair(config.ClientCertificateFile, config.ClientKeyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
	}
	return credentials.NewTLS(tlsConfig), nil
}

func NewWorkerClient(config Configuration) (proto.WorkerServiceClient, error) {
	// tlsCredentials, err := loadTLSCredentials(config)
	// if err != nil {
	// 	return nil, err
	// }
	conn, err := grpc.Dial(
		config.ServerEndpoint,
		//grpc.WithTransportCredentials(tlsCredentials),
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}
	return proto.NewWorkerServiceClient(conn), nil
}
