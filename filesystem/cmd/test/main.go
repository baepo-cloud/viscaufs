package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Simple debug node implementation
type DebugNode struct {
	fs.Inode
}

func (d *DebugNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	fmt.Fprintln(os.Stderr, "DEBUG: Getattr called")
	out.Mode = 0755 | syscall.S_IFDIR
	return 0
}

func (d *DebugNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {
	fmt.Fprintln(os.Stderr, "DEBUG: Readdir called")
	return fs.NewListDirStream(nil), 0
}

func main() {
	mountPoint := flag.String("mount", "/tmp/fuse-debug", "Mount point")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "Mounting at %s\n", *mountPoint)

	root := &DebugNode{}

	opts := &fs.Options{
		MountOptions: fuse.MountOptions{
			Debug: true,
			Name:  "debug-fs",
		},
	}

	server, err := fs.Mount(*mountPoint, root, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Mount failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintln(os.Stderr, "Mounted successfully")

	// Wait for signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Fprintln(os.Stderr, "Unmounting...")
	server.Unmount()
}
