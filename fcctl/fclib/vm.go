package fclib

import (
	"encoding/json"
	"fmt"
	"os"

	"fcctl/util"

	"github.com/firecracker-microvm/firecracker-go-sdk"
)

type (
	VmConf struct {
		Name string `json:"_name,omitempty"`
		Sock string `json:"_sock,omitempty"`
		Log  string `json:"_log,omitempty"`
		File string `json:"-"`
	}
)

func NewVm(name string) *VmConf {
	return &VmConf{
		Name: name,
		Sock: fmt.Sprintf("%s/%s.sock", DefaultRundir, name),
		Log:  fmt.Sprintf("%s/%s.log", DefaultLogdir, name),
		File: fmt.Sprintf("%s/%s.json", DefaultConfdir, name),
	}
}

func (c *VmConf) GetShares() []*ShareConf {
	vm, _ := c.ReadVm()
	shares := make([]*ShareConf, 0)
	if vm == nil {
		return shares
	}
	for _, dev := range vm.VHostUserDevices {
		shares = append(shares, NewShare(dev.ID))
	}
	return shares
}

func (c *VmConf) CheckFile() error {
	return checkFile(c.File, "file", ErrConfigExists, nil)
}

func (c *VmConf) CheckSock() error {
	return checkFile(c.Sock, "sock", ErrSocketExists, nil)
}

func (c *VmConf) GetVMCommandBuilder(firecrackerBin string) firecracker.VMCommandBuilder {
	return firecracker.VMCommandBuilder{}.
		WithBin(firecrackerBin).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithSocketPath(c.Sock).
		//"--config-file", c.File,
		WithArgs([]string{"--id", c.Name})
}

func (c *VmConf) WriteVm(vm *util.VmConfFile) error {
	if err := c.CheckFile(); err != nil {
		return err
	}
	out, err := os.OpenFile(c.File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&vm); err != nil {
		return err
	}
	return nil
}

func (c *VmConf) ReadVm() (*util.VmConfFile, error) {
	f, err := os.Open(c.File)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	conf := &util.VmConfFile{}
	if err := json.NewDecoder(f).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *VmConf) Clean() error {
	// try to cleanup socket
	if err := cleanupFile(c.Sock); err != nil {
		return err
	}
	if vm, _ := c.ReadVm(); vm != nil {
		if vm.Logger != nil && vm.Logger.LogPath != nil {
			// cleanup log registered in vm config file
			if err := cleanupFile(*vm.Logger.LogPath); err != nil {
				return err
			}
		}
		if vm.Vsock != nil && vm.Vsock.UdsPath != nil {
			// cleanup vsock registered in vm config file
			if err := cleanupFile(*vm.Vsock.UdsPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func cleanupFile(fname string) error {
	if err := os.RemoveAll(fname); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// file probably didn't exist
	}
	return nil
}
