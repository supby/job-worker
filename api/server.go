package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/supby/job-worker/workerlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	workerservicepb "github.com/supby/job-worker/generated/proto"
)

func loadTLSCredentials(conf Configuration) (credentials.TransportCredentials, error) {
	pemClientCA, err := ioutil.ReadFile(conf.CAFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	serverCert, err := tls.LoadX509KeyPair(conf.ServerCertificateFile, conf.ServerKeyFile)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}
	return credentials.NewTLS(config), nil
}

func createServer(config Configuration, cred credentials.TransportCredentials) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", config.Endpoint)
	if err != nil {
		return nil, nil, err
	}
	grpcServer := grpc.NewServer(
		//grpc.Creds(cred),
		grpc.UnaryInterceptor(UnaryAuthInterceptor),
		grpc.StreamInterceptor(StreamAuthInterceptor),
	)

	workerservicepb.RegisterWorkerServiceServer(grpcServer, &workerServer{
		Worker: workerlib.New(),
	})
	return grpcServer, lis, nil
}

func StartServer(config Configuration) error {
	cred, err := loadTLSCredentials(config)
	if err != nil {
		log.Printf("Error loading certificates: %v.\n", err)
		return err
	}
	log.Println("Certificates are loaded.")

	serv, lis, err := createServer(config, cred)
	if err != nil {
		log.Printf("Error creating server: %v.\n", err)
		return err
	}
	defer lis.Close()
	log.Println("Listening server is created.")

	log.Printf("Start serving on %v.\n", config.Endpoint)
	if err := serv.Serve(lis); err != nil {
		log.Printf("Error serving: %v.\n", err)
		return err
	}
	return nil
}
