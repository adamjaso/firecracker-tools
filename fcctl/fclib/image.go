package fclib

import "fmt"

type (
	ImageConf struct {
		Name string
		File string
	}
)

func NewImage(name string) *ImageConf {
	return &ImageConf{
		Name: name,
		File: fmt.Sprintf("%s/%s.img.gz", DefaultImagedir, name),
	}
}

func (i *ImageConf) CheckFile() error {
	return checkFile(i.File, "image", ErrFileExists, nil)
}
