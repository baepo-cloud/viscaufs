package types

import "github.com/pkg/errors"

var (
	ErrImageDownloadAlreadyAcquired = errors.New("image download already acquired")
	ErrImageAlreadyPresent          = errors.New("image already present")
)
