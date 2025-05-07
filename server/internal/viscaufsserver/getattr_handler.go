package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s Server) GetAttr(_ context.Context, request *fspb.GetAttrRequest) (*fspb.GetAttrResponse, error) {
	lookup := s.FSIndexerService.Lookup(request.ImageDigest, request.Path)
	if lookup == nil {
		return nil, status.Error(codes.NotFound, "path not found")
	}

	proto := lookup.ToProto()
	return &fspb.GetAttrResponse{
		Attributes: proto.Attributes,
	}, nil
}
