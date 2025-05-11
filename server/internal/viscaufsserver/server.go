package viscaufsserver

import (
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
)

type Server struct {
	ImageService       types.ImageService
	FSIndexerService   types.FileSystemIndexService
	FileHandlerService types.FileHandlerService

	fspb.UnimplementedFuseServiceServer
}

var _ fspb.FuseServiceServer = (*Server)(nil)

func New(imageService types.ImageService, fsIndexerService types.FileSystemIndexService, fhService types.FileHandlerService) *Server {
	return &Server{
		ImageService:       imageService,
		FSIndexerService:   fsIndexerService,
		FileHandlerService: fhService,
	}
}
