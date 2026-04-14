package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

var ErrInvalidTapDevice = errors.New("invalid tap device")

type (
	VmOpts struct {
		nameF                           string
		Vcpu, Mem                       int64
		TapDevice, DiskPath             string
		LogPath                         string
		KernelPath, InitrdPath, Cmdline string
		MetricsPath                     string
		EnableMMDS                      bool
	}
)

func BuildVm(opts VmOpts) (*models.FullVMConfiguration, error) {
	if _, err := os.Stat(opts.DiskPath); err != nil {
		return nil, fmt.Errorf("root disk not found: %w", err)
	}
	kernelF, _ := filepath.Abs(opts.KernelPath)
	diskF, _ := filepath.Abs(opts.DiskPath)
	logF, _ := filepath.Abs(opts.LogPath)
	cfg := models.FullVMConfiguration{
		BootSource: &models.BootSource{
			KernelImagePath: &kernelF,
			BootArgs:        opts.Cmdline,
		},
		Drives: []*models.Drive{
			{
				DriveID:      &diskF,
				IsRootDevice: new(true),
				IsReadOnly:   new(false),
				PathOnHost:   &diskF,
				IoEngine:     new("Sync"),
			},
		},
		Logger: &models.Logger{
			Level:     new("Info"),
			LogPath:   new(logF),
			ShowLevel: new(true),
		},
		MachineConfig: &models.MachineConfiguration{
			MemSizeMib: new(opts.Mem),
			Smt:        new(true),
			VcpuCount:  new(opts.Vcpu),
		},
	}
	if opts.InitrdPath != "" {
		cfg.BootSource.InitrdPath, _ = filepath.Abs(opts.InitrdPath)
	}
	if opts.TapDevice != "" {
		tapdev := strings.SplitN(opts.TapDevice, "/", 2)
		if len(tapdev) != 2 {
			return nil, fmt.Errorf("%w: tap device must be of the form %q, got %q\n", ErrInvalidTapDevice, "tap/mac", opts.TapDevice)
		}
		if _, err := net.InterfaceByName(tapdev[0]); err != nil {
			return nil, fmt.Errorf("%w: tap device %q not found", ErrInvalidTapDevice, tapdev[0])
		}
		cfg.NetworkInterfaces = []*models.NetworkInterface{
			{
				IfaceID:     new(tapdev[0]),
				HostDevName: new(tapdev[0]),
				GuestMac:    tapdev[1],
			},
		}
		if opts.EnableMMDS {
			cfg.MmdsConfig = &models.MmdsConfig{
				IPV4Address:       new("169.254.169.254"),
				NetworkInterfaces: []string{tapdev[0]},
				Version:           new("V1"),
			}
		}
	}
	if opts.MetricsPath != "" {
		cfg.Metrics = &models.Metrics{
			MetricsPath: new(opts.MetricsPath),
		}
	}
	return &cfg, nil
}
