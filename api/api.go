package api

import (
	"context"

	workerservicepb "github.com/supby/job-worker/api/proto"
	"github.com/supby/job-worker/workerlib"
	"github.com/supby/job-worker/workerlib/job"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type workerServer struct {
	workerservicepb.UnimplementedWorkerServiceServer
	Worker workerlib.Worker
}

func (s *workerServer) Start(ctx context.Context, r *workerservicepb.StartRequest) (*workerservicepb.StartResponse, error) {
	jobID, err := s.Worker.Start(job.Command{Name: r.CommandName, Arguments: r.Arguments})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := workerservicepb.StartResponse{
		JobID: jobID[:],
	}
	return &res, nil
}

func (s *workerServer) Stop(ctx context.Context, r *workerservicepb.StopRequest) (*workerservicepb.StopResponse, error) {
	var jobID [16]byte
	copy(jobID[:], r.JobID)
	err := s.Worker.Stop(jobID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &workerservicepb.StopResponse{}, nil
}

func (s *workerServer) QueryStatus(ctx context.Context, r *workerservicepb.QueryStatusRequest) (*workerservicepb.QueryStatusResponse, error) {
	var jobID [16]byte
	copy(jobID[:], r.JobID)
	jobstatus, err := s.Worker.QueryStatus(jobID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := workerservicepb.QueryStatusResponse{
		ExitCode:    int32(jobstatus.ExitCode),
		JobStatus:   workerservicepb.JobStatus(jobstatus.StatusCode),
		CommandName: jobstatus.CommandName,
		Arguments:   jobstatus.Arguments,
	}
	return &res, nil
}

func (s *workerServer) GetOutput(r *workerservicepb.GetOutputRequest, stream workerservicepb.WorkerService_GetOutputServer) error {
	var jobID [16]byte
	copy(jobID[:], r.JobID)
	logchan, err := s.Worker.GetStream(stream.Context(), jobID)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case log, ok := <-logchan:
			if !ok {
				return nil
			}
			if err := stream.SendMsg(&workerservicepb.GetOutputResponse{Output: log}); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}
}
