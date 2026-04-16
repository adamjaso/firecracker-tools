package fclib

import (
	"context"
	"fmt"

	"fcctl/util"
)

type (
	DiskConf struct {
		Name string
		File string
	}
)

func NewDisk(name string) *DiskConf {
	return &DiskConf{
		Name: name,
		File: fmt.Sprintf("%s/%s.img", DefaultDiskdir, name),
	}
}

func (c *DiskConf) CheckFile() error {
	return checkFile(c.File, "disk", ErrFileExists, nil)
}

func (c *DiskConf) Exists() error {
	return checkFile(c.File, "disk", nil, ErrFileNotFound)
}

func (c *DiskConf) InstallFromImage(ctx context.Context, img *ImageConf) error {
	return util.ExecCompress(ctx, "-dc", img.File, c.File)
}
