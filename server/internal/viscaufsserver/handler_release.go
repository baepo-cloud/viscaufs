package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) Release(_ context.Context, req *fspb.ReleaseRequest) (*fspb.ReleaseResponse, error) {
	err := s.FileHandlerService.ReleaseFile(req.Uid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to release file: %v", err)
	}

	return &fspb.ReleaseResponse{}, nil
}
