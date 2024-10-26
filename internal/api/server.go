package api

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	workerservicepb "github.com/supby/job-worker/generated/proto"
	"github.com/supby/job-worker/internal/workerlib"
)

func loadTLSCredentials(conf *Configuration) (credentials.TransportCredentials, error) {
	pemClientCA, err := ioutil.ReadFile(conf.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA file: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	serverCert, err := tls.LoadX509KeyPair(conf.ServerCertificateFile, conf.ServerKeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server key pair: %w", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
		// CipherSuites: []uint16{
		// 	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		// 	tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		// },
	}
	return credentials.NewTLS(config), nil
}

func createServer(config *Configuration, cred credentials.TransportCredentials) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", config.Endpoint)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to listen: %w", err)
	}

	opts := []grpc.ServerOption{
		grpc.Creds(cred),
		grpc.UnaryInterceptor(UnaryAuthInterceptor),
		grpc.StreamInterceptor(StreamAuthInterceptor),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			Time:              10 * time.Second,
			Timeout:           1 * time.Second,
		}),
	}

	grpcServer := grpc.NewServer(opts...)

	workerservicepb.RegisterWorkerServiceServer(grpcServer, NewWorkerServer(workerlib.New()))
	return grpcServer, lis, nil
}

func StartServer(config *Configuration) error {
	cred, err := loadTLSCredentials(config)
	if err != nil {
		return fmt.Errorf("failed to load TLS credentials: %w", err)
	}
	log.Println("TLS credentials loaded successfully")

	serv, lis, err := createServer(config, cred)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}
	defer lis.Close()
	log.Printf("Server created and listening on %s", config.Endpoint)

	go func() {
		log.Printf("Starting to serve on %s", config.Endpoint)
		if err := serv.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")
	serv.GracefulStop()
	log.Println("Server stopped")

	return nil
}
