package types

import "context"

type (
	OpenFileParams struct {
		Path        string
		ImageDigest string
		Flags       uint32
	}
	FileHandlerService interface {
		OpenFile(ctx context.Context, params OpenFileParams) (string, error)
		ReleaseFile(uid string) error
		ReadFile(uid string, offset int64, length uint32) ([]byte, error)
	}
)
