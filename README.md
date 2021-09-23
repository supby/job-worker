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

Standalone application provides CLI interface to communicate with GRPC API over network.

## Security

Transport security is based on TLS 1.3. The cipher suites is: TLS_AES_256_GCM_SHA384.

### Authentification

Authentification is based on x.509 certificates. Server and Client are shared common CA root certificate. Every side checks others side certificate against common CA.

### Authorization

[TBD]

