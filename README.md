# job-worker

The project aims to provide tools to run and control jobs (processes) on remote or local host though GRPC API. It consists of three main parts: The Library, GRPC API and standalone CLI client.

## Overview

### The Library

It is Golang package which provides abstration to control host processes. It supports four operations: 
- Run process
- Stop process. SIGTERM signal should be sent to process.
- Query current process' status
- Streaming process' output(stdout and stderr).


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
    bool success = 2;
    string message = 3;
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
    STOPPED = 3;
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

### Authentification

Authentification is based on x.509 certificates. Every side checks others side certificate against common CA. Clients certificate should be generated and signed by CA during provisioning process.

### Authorization

Client's role should be stored in X.509 v3 extensions of clients certificate. For role storing will be used appropriate extension with OID=1.2.840.10070.8.1. OID reference here http://oid-info.com/get/1.2.840.10070.8.1
Provisioning center generates clients certificate based on clients registration data and assigned role. Using this approach clients certificate can be mapped to appropriate role on server side.

Server should supports two roles:
- Readonly: quering job status, stream jobs output.
- Full: full access to functionality provided by API.


## Trade-offs

### Authorization

Provisioning of clients is not a part of the task. CA and certificates will be generated manually using openssl. Roles will be hardcoded in on server side.

### Configuration

All configuration(server and client) will be hardcoded in app because of simplicity.

### Loggining

Persistent logging system is not a part of the task. Server it self will log in standart output. Logs from jobs processes will be stored in memory.

