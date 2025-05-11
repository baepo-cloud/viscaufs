package viscaufsserver

import (
	"context"
	"fmt"

	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"google.golang.org/grpc/status"
)

func (s Server) Read(ctx context.Context, request *fspb.ReadRequest) (*fspb.ReadResponse, error) {
	bytes, err := s.FileHandlerService.ReadFile(request.Uid, request.Offset, request.Size)
	if err != nil {
		return nil, status.Error(status.Code(err), fmt.Errorf("failed to read file: %v", err).Error())
	}

	return &fspb.ReadResponse{
		Data: bytes,
	}, nil
}
