package util

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	// VHostUserDeviceTypes: https://docs.oasis-open.org/virtio/virtio/v1.3/csd01/virtio-v1.3-csd01.html#x1-1930005
	VHostUserDeviceTypeVirtioSCSI = 8
	VHostUserDeviceTypeVirtioFS   = 26
)

var ErrInvalidTapDevice = errors.New("invalid tap device")

type (
	VHostUserDevice struct {
		ID         string `json:"id"`
		DeviceType int    `json:"device_type"`
		Socket     string `json:"socket"`
		NumQueues  int    `json:"num_queues,omitempty"`
		QueueSize  int    `json:"queue_size,omitempty"`
	}
)

type (
	VmOpts struct {
		nameF                           string
		Vcpu, Mem                       int64
		TapDevice, DiskPath             string
		LogPath                         string
		KernelPath, InitrdPath, Cmdline string
		MetricsPath                     string
		EnableMMDS                      bool
		VirtiofsID                      string
		VirtiofsSock                    string
	}
	// VmConfFile represents a modified config file that supports the non-mainline virtiofsd feature
	// The feature is tracked in https://github.com/firecracker-microvm/firecracker/pull/5773
	// If/once the feature is merged, this can be refactored to just use FullVMConfiguration.
	VmConfFile struct {
		models.FullVMConfiguration
		VHostUserDevices []VHostUserDevice `json:"vhost-user-devices,omitempty"`
		CreatedAt        int64             `json:"_created_at"` // used for sorting
	}
)

func NewVirtiofsDevice(id, sock string) VHostUserDevice {
	return VHostUserDevice{
		ID:         id,
		DeviceType: VHostUserDeviceTypeVirtioFS,
		Socket:     sock,
		NumQueues:  2,
		QueueSize:  256,
	}
}

func BuildVm(opts VmOpts) (*VmConfFile, error) {
	if _, err := os.Stat(opts.DiskPath); err != nil {
		return nil, fmt.Errorf("root disk not found: %w", err)
	}
	kernelF, _ := filepath.Abs(opts.KernelPath)
	diskF, _ := filepath.Abs(opts.DiskPath)
	logF, _ := filepath.Abs(opts.LogPath)
	cfg := VmConfFile{
		CreatedAt: time.Now().Unix(),
		FullVMConfiguration: models.FullVMConfiguration{
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
	if opts.VirtiofsID != "" && opts.VirtiofsSock != "" {
		cfg.VHostUserDevices = []VHostUserDevice{
			NewVirtiofsDevice(opts.VirtiofsID, opts.VirtiofsSock),
		}
	}
	return &cfg, nil
}
