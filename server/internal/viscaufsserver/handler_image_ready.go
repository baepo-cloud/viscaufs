package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) ImageReady(_ context.Context, request *fspb.ImageReadyRequest) (*fspb.ImageReadyResponse, error) {
	ready := s.FSIndexerService.Ready(request.ImageDigest)
	if !ready {
		return nil, status.Error(codes.FailedPrecondition, "image not ready")
	}

	return &fspb.ImageReadyResponse{}, nil
}
