package fclib

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	DefaultConfdir   = "/var/lib/firecracker/conf"
	DefaultDiskdir   = "/var/lib/firecracker/disk"
	DefaultKerneldir = "/var/lib/firecracker/kernel"
	DefaultLogdir    = "/var/log/firecracker"
	DefaultRundir    = "/var/run/firecracker"
)

var (
	dirs = []string{
		DefaultConfdir,
		DefaultDiskdir,
		DefaultKerneldir,
		DefaultLogdir,
		DefaultRundir,
	}

	ErrConfigExists   = errors.New("config exists")
	ErrSocketExists   = errors.New("socket exists")
	ErrKernelNotFound = errors.New("kernel not found")
	ErrInitrdNotFound = errors.New("initrd not found")
	ErrRootfsExists   = errors.New("tarball exists")
	ErrChrootExists   = errors.New("script exists")
	ErrDiskExists     = errors.New("file exists")
)

type (
	Conf struct {
		Name string
		Sock string
		Log  string
		File string
	}
)

func InitConfdirs() error {
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func checkFile(file, name string, existErr, notExistErr error) error {
	if _, err := os.Stat(file); err == nil {
		if existErr != nil {
			return fmt.Errorf("%w: %s %s", existErr, name, file)
		}
		return nil
	} else if !os.IsNotExist(err) {
		return err
	} else if notExistErr != nil {
		return notExistErr
	}
	return nil
}

func New(name string) *Conf {
	return &Conf{
		Name: name,
		Sock: fmt.Sprintf("%s/%s.sock", DefaultRundir, name),
		Log:  fmt.Sprintf("%s/%s.log", DefaultLogdir, name),
		File: fmt.Sprintf("%s/%s.json", DefaultConfdir, name),
	}
}

func (c *Conf) CheckFile() error {
	return checkFile(c.File, "file", ErrConfigExists, nil)
}

func (c *Conf) CheckSock() error {
	return checkFile(c.Sock, "sock", ErrSocketExists, nil)
}

func (c *Conf) GetVMCommandBuilder(firecrackerBin string) firecracker.VMCommandBuilder {
	return firecracker.VMCommandBuilder{}.
		WithBin(firecrackerBin).
		WithStdin(os.Stdin).
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithSocketPath(c.Sock).
		WithArgs([]string{"--config-file", c.File, "--id", c.Name, "--log-path", c.Log})
}

func (c *Conf) WriteVm(vm *models.FullVMConfiguration) error {
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

func (c *Conf) ReadVm() (*models.FullVMConfiguration, error) {
	f, err := os.Open(c.File)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	conf := &models.FullVMConfiguration{}
	if err := json.NewDecoder(f).Decode(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func (c *Conf) Clean() error {
	// try to cleanup socket
	if err := cleanupFile(c.Sock); err != nil {
		return err
	}
	// try to cleanup log
	if err := cleanupFile(c.Log); err != nil {
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
	if err := os.Remove(fname); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// file probably didn't exist
	}
	return nil
}
