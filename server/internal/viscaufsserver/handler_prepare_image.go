package viscaufsserver

import (
	"context"
	"errors"
	"log/slog"

	"github.com/baepo-cloud/viscaufs-server/internal/types"
	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) PrepareImage(_ context.Context, request *fspb.PrepareImageRequest) (*fspb.PrepareImageResponse, error) {
	digest, err := s.ImageService.Download(request.ImageRef)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrImageAlreadyPresent), errors.Is(err, types.ErrImageDownloadAlreadyAcquired):
		default:
			slog.Error("unable to retrieve image", "error", err)
			return nil, status.Error(codes.Internal, "unable to download image")
		}
	}

	return &fspb.PrepareImageResponse{
		ImageDigest: digest,
	}, nil
}
