package viscaufsserver

import (
	"context"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
)

func (s Server) ReadDir(ctx context.Context, request *fspb.ReadDirRequest) (*fspb.ReadDirResponse, error) {
	entriesByPrefix := s.FSIndexerService.LookupByPrefix(request.ImageDigest, request.Path)
	var entries []*fspb.File
	for _, entry := range entriesByPrefix {
		e := entry.ToProto()
		entries = append(entries, &fspb.File{
			Path:          entry.Path,
			Attributes:    e.Attributes,
			SymlinkTarget: e.SymlinkTarget,
		})
	}

	return &fspb.ReadDirResponse{
		Entries: entries,
	}, nil
}
