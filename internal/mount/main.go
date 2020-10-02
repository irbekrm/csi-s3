package mount

import "fmt"

func NewMounter(mounter, mounterBinaryPath string) (Mounter, error) {
	switch mounter {
	case "s3fs":
		return S3fs{path: mounterBinaryPath}, nil
	default:
		return nil, fmt.Errorf("unknow mounter: %s", mounter)
	}
}

type Mounter interface {
	IsReady() (bool, error)
}

type S3fs struct {
	path string
}

func (s S3fs) IsReady() (bool, error) {
	return true, nil
}
