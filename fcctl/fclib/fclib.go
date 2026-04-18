package fclib

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfdir   = "/var/lib/firecracker/conf"
	DefaultDiskdir   = "/var/lib/firecracker/disk"
	DefaultKerneldir = "/var/lib/firecracker/kernel"
	DefaultSharedir  = "/var/lib/firecracker/share"
	DefaultImagedir  = "/var/lib/firecracker/image"
	DefaultNetdir    = "/var/lib/firecracker/net"
	DefaultLogdir    = "/var/log/firecracker"
	DefaultRundir    = "/var/run/firecracker"
)

var (
	dirs = []string{
		DefaultConfdir,
		DefaultDiskdir,
		DefaultKerneldir,
		DefaultSharedir,
		DefaultImagedir,
		DefaultNetdir,
		DefaultLogdir,
		DefaultRundir,
	}

	ErrConfigExists   = errors.New("config exists")
	ErrSocketExists   = errors.New("socket exists")
	ErrKernelNotFound = errors.New("kernel not found")
	ErrInitrdNotFound = errors.New("initrd not found")
	ErrInitrdExists   = errors.New("initrd exists")
	ErrKernelExists   = errors.New("kernel exists")
	ErrRootfsExists   = errors.New("tarball exists")
	ErrChrootExists   = errors.New("script exists")
	ErrFileExists     = errors.New("file exists")
	ErrFileNotFound   = errors.New("file not found")

	Progname       string = filepath.Base(os.Args[0])
	FirecrackerBin string
)

func init() {
	if FirecrackerBin = os.Getenv("FC_BIN"); FirecrackerBin == "" {
		FirecrackerBin = "firecracker"
	}
}

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
		return fmt.Errorf("%w: %s %s", err, name, file)
	} else if notExistErr != nil {
		return notExistErr
	}
	return nil
}
