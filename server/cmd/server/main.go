package main

import (
	"github.com/baepo-cloud/viscaufs-server/internal/config"
	"github.com/baepo-cloud/viscaufs-server/internal/fxutil"
	"github.com/baepo-cloud/viscaufs-server/internal/service/filehandlerservice"
	"github.com/baepo-cloud/viscaufs-server/internal/service/fsindexservice"
	"github.com/baepo-cloud/viscaufs-server/internal/service/imgservice"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"github.com/baepo-cloud/viscaufs-server/internal/viscaufsserver"
	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func main() {
	fx.New(
		fxutil.Logger(),
		fx.Provide(fxutil.ProvideGORM),
		fx.Provide(fxutil.ProvideGRPCServer),
		fx.Provide(config.ParseConfig),
		fx.Provide(fx.Annotate(fsindexservice.NewService, fx.As(new(types.FileSystemIndexService)))),
		fx.Provide(fx.Annotate(imgservice.NewService, fx.As(new(types.ImageService)))),
		fx.Provide(fx.Annotate(filehandlerservice.NewService, fx.As(new(types.FileHandlerService)))),
		fx.Provide(viscaufsserver.New),
		fx.Invoke(func(server *grpc.Server) {}),
		//fx.Invoke(func(db *gorm.DB) {
		//	img := "sha256:86b823a6ef96fb1766da15f65eceb1378b748f45e3ef4ab00c7c7b0d8e00e46b"
		//	var image types.Image
		//	if err := db.Where("digest = ?", img).First(&image).Error; err != nil {
		//		return
		//	}
		//
		//	now := time.Now()
		//	fi, _ := fsindex.Deserialize(image.FsIndex, true)
		//	slog.Info("time deserialize", slog.String("time", time.Since(now).String()))
		//	s := fi.String()
		//
		//	search := fi.LookupPrefixSearch("/")
		//	for _, node := range search {
		//		slog.Info("node", slog.String("node", node.Path))
		//	}
		//
		//	// write into a file
		//	os.WriteFile("fsindex_alpine.txt", []byte(s), 0644)
		//	slog.Info("fsindex created")
		//}),
	).Run()
}
