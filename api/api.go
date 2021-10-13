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
	jobID, err := s.Worker.Start(job.Command{Name: r.CommandName, Args: r.Arguments})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := workerservicepb.StartResponse{
		JobID: jobID,
	}
	return &res, nil
}

func (s *workerServer) Stop(ctx context.Context, r *workerservicepb.StopRequest) (*workerservicepb.StopResponse, error) {
	err := s.Worker.Stop(r.JobID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &workerservicepb.StopResponse{}, nil
}

func (s *workerServer) QueryStatus(ctx context.Context, r *workerservicepb.QueryStatusRequest) (*workerservicepb.QueryStatusResponse, error) {
	jobstatus, err := s.Worker.QueryStatus(r.JobID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	res := workerservicepb.QueryStatusResponse{
		ExitCode: int32(jobstatus.ExitCode),
		Exited:   jobstatus.Exited,
	}
	return &res, nil
}

func (s *workerServer) GetOutput(r *workerservicepb.GetOutputRequest, stream workerservicepb.WorkerService_GetOutputServer) error {
	logchan, err := s.Worker.GetStream(stream.Context(), r.JobID)
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
