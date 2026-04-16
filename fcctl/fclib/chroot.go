package fclib

import "fmt"

type (
	ChrootConf struct {
		Name   string
		Dir    string
		Script string
	}
)

func NewChroot(name string) *ChrootConf {
	return &ChrootConf{
		Name:   name,
		Dir:    fmt.Sprintf("%s/%s.chroot", DefaultRundir, name),
		Script: fmt.Sprintf("%s/%s.chroot.sh", DefaultConfdir, name),
	}
}

func (c *ChrootConf) CheckScript() error {
	return checkFile(c.Script, "chroot script", ErrChrootExists, nil)
}
