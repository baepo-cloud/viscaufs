package types

type (
	ImageService interface {
		Download(refId string) (string, error)
	}
)
