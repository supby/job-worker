# job-worker

The project aims to provide tools to run and control jobs (processes) on remote or local host though GRPC API. It consists of three main parts: The Library, GRPC API and standalone CLI client.

## Overview

### The Library

It is Golang package which provides abstration to control host processes. It supports four operations: 
- Run process
- Stop process. SIGTERM signal should be sent to process.
- Query current process' status
- Streaming process' output(stdout and stderr). On the library level stdout/stderr (io.Writer)  will be assigned with in-memory io.Writer   implemetation which pushes output data to chain(golang chain) on every Write from the process. The chain data will be consumed in Stream method in API server.


### GRPC API

Exposes API to provide access to library's functionality over network. API is responsible for authentification, authorization and TLS communication. Status of API execution should be returned using standart GRPC status codes. Possible status codes:
 - INTERNAL: Some inernal error.
 - NOT_FOUND: Requested jobID is not found. 
 - INVALID_ARGUMENT: Invalid data format is provided. For instance, jobId should be UUID. Command is empty.
 - UNAUTHENTICATED: Client request cannot be authenticated. (no cert, wrong cert)
 - PERMISSION_DENIED: Client request authenticated but doesn't have permission to perform some operation. For instance 'readonly' cannot stop job.

```
syntax = "proto3";

package workerservice;

message StartRequest {
    string commandName = 1;
    repeated string arguments = 2;
}
  
message StartResponse {
    bytes jobID = 1;
}
  
message StopRequest {
    bytes jobID = 1;
}
  
message StopResponse { }
  
message QueryStatusRequest {
    bytes jobID = 1;
}

enum JobStatus {
    UNKNOWN = 0;
    RUNNING = 1;    
    EXITED = 2;
    STOPPED = 3;
    STARTED = 4;
}
  
message QueryStatusResponse {
    int32 exitCode = 1;
    string commandName = 2;
    repeated string arguments = 3;
    JobStatus JobStatus = 4;
}
  
message GetOutputRequest {
    bytes jobID = 1;
}
  
message GetOutputResponse {
    bytes output = 1;
}

service WorkerService {
    rpc Start(StartRequest) returns (StartResponse);
    rpc Stop(StopRequest) returns (StopResponse);
    rpc QueryStatus(QueryStatusRequest) returns (QueryStatusResponse);
    rpc GetOutput(GetOutputRequest) returns (stream GetOutputResponse);
}
```

### CLI client

Standalone application provides CLI interface to communicate with server GRPC API over network.
Usage: 
``` 
workerclient start <command> -args <arg1> <arg2>
workerclient stop|query|stream -j <job_id>

```

Conection related configuration should be in yaml file. (but in due to simplicity it will be hardcoded in app)
```
serverAddress: "localhost:5000"
serverCA: "path to CA file"
clientCertificate: "path to client cert"
clientKey: "path to private key of client cert"
```


## Security

Transport security is based on TLS 1.3. The cipher suites is: TLS_AES_256_GCM_SHA384.

### Authentication

Authentification is based on x.509 certificates. Every side checks others side certificate against common CA. Clients certificate should be generated and signed by CA during provisioning process.

### Authorization

Client's role should be stored in X.509 v3 extensions of clients certificate. For role storing will be used appropriate extension with OID=1.2.840.10070.8.1. OID reference here http://oid-info.com/get/1.2.840.10070.8.1
Provisioning center generates clients certificate based on clients registration data and assigned role. Using this approach clients certificate can be mapped to appropriate role on server side.

Server should supports two roles:
- Readonly: quering job status, stream jobs output.
- Full: full access to functionality provided by API.


## Misc

### Generate test certificates

```
make gentestcert
```
