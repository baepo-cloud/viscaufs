package main

import (
	"github.com/baepo-cloud/viscaufs-server/internal/config"
	"github.com/baepo-cloud/viscaufs-server/internal/fxutil"
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
		fx.Provide(viscaufsserver.New),
		fx.Invoke(func(server *grpc.Server) {}),
	).Run()
}
