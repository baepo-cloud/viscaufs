package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hanwen/go-fuse/v2/fuse"

	"github.com/baepo-cloud/viscaufs-fs/viscaufs"
	fspb "github.com/baepo-cloud/viscaufs/common/proto/gen/v1"
	"github.com/hanwen/go-fuse/v2/fs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Parse command line arguments
	var (
		serverAddr  string
		mountPoint  string
		imageDigest string
		debug       bool
	)

	flag.StringVar(&serverAddr, "server", "localhost:8080", "filesystem server address")
	flag.StringVar(&mountPoint, "mount", "/mnt/viscaufs", "Mount point for FUSE filesystem")
	flag.StringVar(&imageDigest, "digest", "", "Docker image reference ID")
	flag.BoolVar(&debug, "debug", true, "Enable debug logging")
	flag.Parse()

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)

	if imageDigest == "" {
		slog.Error("image reference ID is required")
		os.Exit(1)
	}

	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		slog.Error("failed to connect to gRPC server", "error", err)
		os.Exit(1)
	}
	defer conn.Close()

	client := fspb.NewFuseServiceClient(conn)

	waitForImageReady(client, imageDigest, time.Second*10)

	// Create the filesystem object
	vfs := &viscaufs.FS{
		Client:      client,
		ImageDigest: imageDigest,
		Cache:       viscaufs.NewCache(),
		MountPath:   mountPoint,
	}

	// Create the root node directly
	rootNode := &viscaufs.Node{
		FS:   vfs,
		Path: "/",
	}

	// Setup FUSE options
	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug:                true,
			Name:                 "viscaufs",
			FsName:               "viscaufs",
			RememberInodes:       true,
			EnableSymlinkCaching: true,
			DisableXAttrs:        true,
		},
	}

	// Mount the filesystem
	server, err := fs.Mount(mountPoint, rootNode, opts)
	if err != nil {
		slog.Error("failed to mount filesystem", "error", err)
		os.Exit(1)
	}

	slog.Info("successfully mounted filesystem", "mount_point", mountPoint)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		slog.Info("received interrupt signal, unmounting filesystem")
		server.Unmount()
	}()

	server.Wait()
}

func waitForImageReady(client fspb.FuseServiceClient, ref string, timeout time.Duration) {
	now := time.Now()
	for {
		_, err := client.ImageReady(context.Background(), &fspb.ImageReadyRequest{ImageDigest: ref})
		if err != nil {
			if time.Since(now) > timeout {
				slog.Error("Image not ready", "error", err)
				os.Exit(1)
			}
			time.Sleep(time.Millisecond * 150)
			continue
		}

		break
	}
}
