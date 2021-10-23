package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/supby/job-worker/workerlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	workerservicepb "github.com/supby/job-worker/api/proto"
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
		grpc.Creds(cred),
		grpc.UnaryInterceptor(UnaryAuthInterceptor),
		grpc.StreamInterceptor(StreamAuthInterceptor),
	)

	workerservicepb.RegisterWorkerServiceServer(grpcServer, &workerServer{
		Worker: workerlib.New(),
	})
	return grpcServer, lis, nil
}

func StartServerTLS(config Configuration) error {
	cred, err := loadTLSCredentials(config)
	if err != nil {
		return err
	}
	serv, lis, err := createServer(config, cred)
	if err != nil {
		return err
	}
	defer lis.Close()
	if err := serv.Serve(lis); err != nil {
		return err
	}
	return nil
}

func StartServer(config Configuration) error {
	lis, err := net.Listen("tcp", config.Endpoint)
	defer lis.Close()

	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()

	workerservicepb.RegisterWorkerServiceServer(grpcServer, &workerServer{
		Worker: workerlib.New(),
	})

	grpcServer.Serve(lis)

	return nil
}
