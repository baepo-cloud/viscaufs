package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
)

func (s Server) ReadDir(ctx context.Context, request *fspb.ReadDirRequest) (*fspb.ReadDirResponse, error) {
	/*
		lookup := s.FSIndexerService.Lookup(request.ImageDigest, request.Path)
			if lookup == nil {
				return nil, status.Error(codes.NotFound, "path not found")
			}

			proto := lookup.ToProto()
			return &fspb.GetAttrResponse{
				Attributes: proto.Attributes,
			}, nil
	*/

	entriesByPrefix := s.FSIndexerService.LookupByPrefix(request.ImageDigest, request.Path)
	var entries []*fspb.DirEntry
	for _, entry := range entriesByPrefix {
		entries = append(entries, &fspb.DirEntry{
			Path:       entry.Path,
			Attributes: entry.ToProto().Attributes,
		})
	}

	return &fspb.ReadDirResponse{
		Entries: entries,
	}, nil
}
