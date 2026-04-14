package fclib

import "fmt"

type (
	RootfsConf struct {
		Tarball string
	}
)

func NewRootfs(name string) *RootfsConf {
	return &RootfsConf{
		Tarball: fmt.Sprintf("%s/%s.rootfs.tar.gz", DefaultConfdir, name),
	}
}

func (r *RootfsConf) CheckTarball() error {
	return checkFile(r.Tarball, "rootfs tarball", ErrRootfsExists, nil)
}
