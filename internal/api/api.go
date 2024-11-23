package api

import (
	"context"
	"errors"
	"log"

	workerservicepb "github.com/supby/job-worker/generated/proto"
	"github.com/supby/job-worker/internal/workerlib"
	"github.com/supby/job-worker/internal/workerlib/job"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WorkerServer struct {
	workerservicepb.UnimplementedWorkerServiceServer
	Worker workerlib.Worker
}

func NewWorkerServer(worker workerlib.Worker) *WorkerServer {
	return &WorkerServer{Worker: worker}
}

func (s *WorkerServer) Start(ctx context.Context, r *workerservicepb.StartRequest) (*workerservicepb.StartResponse, error) {
	if r.CommandName == "" {
		return nil, status.Error(codes.InvalidArgument, "command name is required")
	}

	jobID, err := s.Worker.Start(ctx, job.Command{Name: r.CommandName, Arguments: r.Arguments})
	if err != nil {
		log.Printf("[api] failed to start job: %v", err)
		return nil, status.Error(codes.Internal, "failed to start job")
	}

	res := &workerservicepb.StartResponse{
		JobID: jobID[:],
	}
	log.Printf("[api] job started: %x", jobID)

	return res, nil
}

func (s *WorkerServer) Stop(ctx context.Context, r *workerservicepb.StopRequest) (*workerservicepb.StopResponse, error) {
	jobID, err := s.getJobID(r.JobID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid job ID")
	}

	if err := s.Worker.Stop(ctx, jobID); err != nil {
		if errors.Is(err, workerlib.ErrJobNotFound) {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		log.Printf("[api] failed to stop job %x: %v", jobID, err)
		return nil, status.Error(codes.Internal, "failed to stop job")
	}

	log.Printf("[api] job stopped: %x", jobID)
	return &workerservicepb.StopResponse{}, nil
}

func (s *WorkerServer) QueryStatus(ctx context.Context, r *workerservicepb.QueryStatusRequest) (*workerservicepb.QueryStatusResponse, error) {
	jobID, err := s.getJobID(r.JobID)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid job ID")
	}

	jobStatus, err := s.Worker.QueryStatus(ctx, jobID)
	if err != nil {
		if errors.Is(err, workerlib.ErrJobNotFound) {
			return nil, status.Error(codes.NotFound, "job not found")
		}
		log.Printf("[api] failed to query status for job %x: %v", jobID, err)
		return nil, status.Error(codes.Internal, "failed to query job status")
	}

	return &workerservicepb.QueryStatusResponse{
		ExitCode:    int32(jobStatus.ExitCode),
		JobStatus:   workerservicepb.JobStatus(jobStatus.StatusCode),
		CommandName: jobStatus.CommandName,
		Arguments:   jobStatus.Arguments,
	}, nil
}

func (s *WorkerServer) GetOutput(r *workerservicepb.GetOutputRequest, stream workerservicepb.WorkerService_GetOutputServer) error {
	jobID, err := s.getJobID(r.JobID)
	if err != nil {
		return status.Error(codes.InvalidArgument, "invalid job ID")
	}

	logChan, err := s.Worker.GetStream(stream.Context(), jobID)
	if err != nil {
		if errors.Is(err, workerlib.ErrJobNotFound) {
			return status.Error(codes.NotFound, "job not found")
		}
		log.Printf("[api] failed to get stream for job %x: %v", jobID, err)
		return status.Error(codes.Internal, "failed to get job output stream")
	}

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case logData, ok := <-logChan:
			if !ok {
				return nil
			}
			res := &workerservicepb.GetOutputResponse{Output: logData}
			if err := stream.Send(res); err != nil {
				log.Printf("[api] failed to send output for job %x: %v", jobID, err)
				return status.Error(codes.Internal, "failed to send job output")
			}
		}
	}
}

func (s *WorkerServer) getJobID(j []byte) (job.JobID, error) {
	if len(j) != 16 {
		return job.JobID{}, errors.New("invalid job ID length")
	}
	var jobID job.JobID
	copy(jobID[:], j)
	return jobID, nil
}
