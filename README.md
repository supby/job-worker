# job-worker

The project aims to provide tools to run and control jobs (processes) on remote or local host though GRPC API. It consists of three main parts: The Library, GRPC API and standalone CLI client.

## Overview

### The Library

It is Golang package which provides abstration to control host processes. It supports four operations: 
- Run process
- Stop process
- Query current process' status
- Streaming process' output.


### GRPC API

Exposes API to provide access to library's functionality over network. API is responsible for authentification, authorization and TLS communication.

```
syntax = "proto3";

package workerservice;

message StartRequest {
    string commandName = 1;
    repeated string Arguments = 2;
}
  
message StartResponse {
    string jobID = 1;
    int32 pid = 2;
    bool success = 3;
    string message = 4;
}
  
message StopRequest {
    string jobID = 1;
}
  
message StopResponse {    
    bool success = 1;
    string message = 2;
}
  
message QueryStatusRequest {
    string jobID = 1;
}

enum JobStatus {
    UNKNOWN = 0;
    RUNNING = 1;    
    EXITED = 2;
}
  
message QueryStatusResponse {
    int32 pid = 1;
    int32 exitCode = 2;
    string command = 3;
    JobStatus JobStatus = 4;
}
  
message GetOutputRequest {
    string jobID = 1;
}
  
message GetOutputResponse {
    string output = 1;
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

## Security

Transport security is based on TLS 1.3. The cipher suites is: TLS_AES_256_GCM_SHA384.

### Authentification

Authentification is based on x.509 certificates. Every side checks others side certificate against common CA. Clients certificate should be generated and signed by CA during provisioning process.

### Authorization

Client's role should be stored in X.509 v3 extensions of clients certificate. Provisioning center generates clients certificate based on clients registration data and assigned role. Using this approach clients certificate can be mapped to appropriate role on server side.

Server should supports two roles:
- Readonly: quering job status, stream job output.
- Full: full access to functionality provided by API.


## Trade-offs

### Authorization

As provisioning of client is not part of the task, CA and certificates will be generated manually using openssl. Roles will be hardcoded in memory on server side.

### Loggining

As persistent logging system is not part of the task. Server it self will log in standart output. Logs from jobs processes will be stored in memory only with some rotation based on size.

