test:
	go test ./...

api:
	go build -o ./bin/worker-api cli/apiserver/main.go

client:
	go build -o ./bin/client-cli cli/client/main.go
	
proto:
	rm -rf generated \
	&& mkdir generated \
	&& protoc --go_out=generated --go_opt=paths=source_relative --go-grpc_out=generated --go-grpc_opt=paths=source_relative proto/workerservice.proto