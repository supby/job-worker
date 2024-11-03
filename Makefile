#.PHONY : genproto test buildapi buildclient gentestcert

test:
	go test ./...

buildapi:
	go build -o ./bin/worker-api cmd/apiserver/main.go

buildclient:
	go build -o ./bin/client-cli cmd/client/main.go

genproto:
	rm -rf generated \
	&& mkdir generated \
	&& docker run --rm -v $(shell pwd):/workspace -w /workspace --user 1000 namely/protoc-all:1.51_2 -d proto -l go -o generated/proto --go-source-relative


openssl = docker run -ti --rm -v $(shell pwd)/cert:/apps -w /apps alpine/openssl

_gencnf:
	echo "subjectAltName=DNS:localhost" > $(shell pwd)/cert/openssl.cnf 

_gentestca:
	$(openssl) genrsa -des3 -out rootCA.key -passout pass:testca 2048 \
	&& $(openssl) req -x509 -new -nodes -key rootCA.key -subj '/CN=localhost/O=test./C=US' -passin pass:testca -sha256 -days 1825 -out rootCA.pem

_gettestservercert:
	$(openssl) genrsa -out server.key  -passout pass:testserver 2048 \
	&& $(openssl) req -new -key server.key -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost" -out server.csr \
	&& $(openssl) x509 -req -extfile openssl.cnf -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -passin pass:testca -out server.crt -days 825 -sha256

_gettestclientcert:
	$(openssl) genrsa -out client.key  -passout pass:testclient 2048 \
	&& $(openssl) req -new -key client.key -subj "/CN=localhost" -addext "subjectAltName=DNS:localhost" -out client.csr \
	&& $(openssl) x509 -req -in client.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -passin pass:testca -out client.crt -days 825 -sha256

gentestcert: _gencnf _gentestca _gettestservercert _gettestclientcert

all: genproto test buildapi buildclient