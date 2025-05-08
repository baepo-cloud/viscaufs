package viscaufsserver

import (
	"context"
	"errors"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) Open(ctx context.Context, request *fspb.OpenRequest) (*fspb.OpenResponse, error) {
	uid, err := s.FileHandlerService.OpenFile(ctx, types.OpenFileParams{
		Path:        request.Path,
		ImageDigest: request.ImageDigest,
		Flags:       request.Flags,
	})

	if err != nil {
		switch {
		case errors.Is(err, types.ErrFileNotFound):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &fspb.OpenResponse{
		Uid: uid,
	}, nil
}
