package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	ImageService     types.ImageService
	FSIndexerService types.FileSystemIndexService

	fspb.UnimplementedFuseServiceServer
}

var _ fspb.FuseServiceServer = (*Server)(nil)

func New(imageService types.ImageService, fsIndexerService types.FileSystemIndexService) *Server {
	return &Server{
		ImageService:     imageService,
		FSIndexerService: fsIndexerService,
	}
}

func (s Server) ImageReady(_ context.Context, request *fspb.ImageReadyRequest) (*fspb.ImageReadyResponse, error) {
	ready := s.FSIndexerService.Ready(request.ImageDigest)
	if !ready {
		return nil, status.Error(codes.FailedPrecondition, "image not ready")
	}

	return &fspb.ImageReadyResponse{}, nil
}

func (s Server) Open(ctx context.Context, request *fspb.OpenRequest) (*fspb.OpenResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s Server) Read(ctx context.Context, request *fspb.ReadRequest) (*fspb.ReadResponse, error) {
	//TODO implement me
	panic("implement me")
}
