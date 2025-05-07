package fxutil

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"

	fspb "github.com/baepo-cloud/viscaufs-common/proto/gen/v1"
	"github.com/baepo-cloud/viscaufs-server/db/migrations"
	"github.com/baepo-cloud/viscaufs-server/internal/config"
	"github.com/baepo-cloud/viscaufs-server/internal/types"
	"github.com/baepo-cloud/viscaufs-server/internal/viscaufsserver"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"gorm.io/driver/sqlite"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ProvideGRPCServer(lc fx.Lifecycle, cfg *config.Config, server *viscaufsserver.Server) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(logging.UnaryServerInterceptor(interceptorLogger(slog.Default()))),
		grpc.ChainStreamInterceptor(logging.StreamServerInterceptor(interceptorLogger(slog.Default()))))
	fspb.RegisterFuseServiceServer(grpcServer, server)

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", cfg.Addr)
			if err != nil {
				return err
			}

			slog.Info("grpc server starting", slog.String("addr", cfg.Addr), slog.String("service", "grpc"))
			go grpcServer.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("grpc server shutting down", slog.String("service", "grpc"))
			grpcServer.GracefulStop()
			return nil
		},
	})

	return grpcServer
}

func ProvideGORM(cfg *config.Config) (*gorm.DB, error) {
	// Print the current working directory and SQLite directory
	currentDir, _ := os.Getwd()

	// Create absolute path if needed
	sqliteDir := cfg.SqliteDir
	if !filepath.IsAbs(sqliteDir) {
		sqliteDir = filepath.Join(currentDir, sqliteDir)
	}

	// Ensure the parent directory exists
	if err := os.MkdirAll(sqliteDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create sqlite dir: %w", err)
	}

	sqlitePathFile := filepath.Join(sqliteDir, "storage.db")

	// Create the file if it doesn't exist
	file, err := os.OpenFile(sqlitePathFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create sqlite file: %w", err)
	}
	file.Close()

	// Use direct file path for SQLite
	dbURI := sqlitePathFile

	migrationClient, err := migrations.NewMigrations("sqlite:/" + sqlitePathFile)
	if err != nil {
		return nil, fmt.Errorf("migration client creation failed: %w", err)
	}

	slog.Info("running migrations", slog.String("db_uri", dbURI), slog.String("service", "migration"))
	if err = migrationClient.Migrate(); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbURI), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	err = db.SetupJoinTable(&types.Image{}, "Layers", &types.ImageLayer{})
	if err != nil {
		return nil, fmt.Errorf("failed to setup join table: %w", err)
	}

	err = db.SetupJoinTable(&types.Layer{}, "Images", &types.ImageLayer{})
	if err != nil {
		return nil, fmt.Errorf("failed to setup join table: %w", err)
	}

	return db, nil
}

// interceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
