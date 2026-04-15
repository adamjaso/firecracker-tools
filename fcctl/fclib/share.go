package fclib

import (
	"context"
	"errors"
	"fmt"

	"fcctl/util"
)

type (
	ShareConf struct {
		Name string
		Dir  string
	}
)

func NewShare(name string) *ShareConf {
	return &ShareConf{
		Name: name,
		Dir:  fmt.Sprintf("%s/%s.share", DefaultSharedir, name),
	}
}

func (c *ShareConf) GetSock(sockid string, clean bool) (string, error) {
	sock := fmt.Sprintf("%s/%s.%s.share.sock", DefaultRundir, c.Name, sockid)
	err := checkFile(sock, fmt.Sprintf("sharing %q with %q", c.Name, sockid), ErrFileExists, nil)
	if clean && errors.Is(err, ErrFileExists) {
		if err := cleanupFile(sock); err != nil {
			return sock, err
		}
	}
	return sock, err
}

func (c *ShareConf) GetVirtiofsDevice(sockid string) util.VHostUserDevice {
	sock, _ := c.GetSock(sockid, false)
	return util.NewVirtiofsDevice(sockid, sock)
}

func (c *ShareConf) CheckDir() error {
	return checkFile(c.Dir, "share dir", ErrFileExists, nil)
}

func (c *ShareConf) Exec(ctx context.Context, sockid string, clean bool) error {
	if sock, err := c.GetSock(sockid, clean); err != nil {
		return err
	} else if err := util.RunVirtiofsd(ctx, "", sock, c.Name, c.Dir, ""); err != nil {
		return err
	}
	return nil
}
