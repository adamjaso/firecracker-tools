package fclib

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

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
	if err := util.CheckUnixSocket(sock); err != nil {
		if !errors.Is(err, util.ErrSocket) {
			return sock, err
		}
		if err := cleanupFile(sock); err != nil {
			return sock, err
		}
		return sock, nil
	}
	return sock, nil
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
	} else if cmd, err := util.StartVirtiofsd(ctx, "", sock, c.Name, c.Dir, ""); err != nil {
		return err
	} else {
		util.AssertNoErr(cmd.Wait())
	}
	return nil
}

func (c *ShareConf) Start(ctx context.Context, sockid string, clean bool) (*exec.Cmd, error) {
	if sock, err := c.GetSock(sockid, clean); err != nil {
		return nil, err
	} else if cmd, err := util.StartVirtiofsd(ctx, "", sock, c.Name, c.Dir, ""); err != nil {
		return nil, err
	} else {
		return cmd, nil
	}
}
