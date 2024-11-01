.PHONY : proto test api client

test:
	go test ./...

api:
	go build -o ./bin/worker-api cmd/apiserver/main.go

client:
	go build -o ./bin/client-cli cmd/client/main.go

colon = :

proto:
	rm -rf generated \
	&& mkdir generated \
	&& docker run --rm -v $(shell pwd):/workspace -p 444:444 --user 1000 -w /workspace namely/protoc-all:1.51_2 -d proto -l go -o generated/proto --go-source-relative --with-validator

all: proto test api client