package fclib

import "fmt"

type (
	DiskConf struct {
		File string
	}
)

func NewDisk(name string) *DiskConf {
	return &DiskConf{
		File: fmt.Sprintf("%s/%s.img", DefaultDiskdir, name),
	}
}

func (c *DiskConf) CheckFile() error {
	return checkFile(c.File, "disk image", ErrDiskExists, nil)
}
