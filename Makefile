test:
	go test ./...

api:
	go build -o ./bin/worker-api main.go
	
proto:
	protoc --go_out=api --go_opt=paths=source_relative --go-grpc_out=api --go-grpc_opt=paths=source_relative proto/workerservice.proto