package fclib

import (
	"fmt"
)

type (
	KernelConf struct {
		Vmlinux string
		Initrd  string
		Config  string
	}
)

func NewKernel(name string) *KernelConf {
	return &KernelConf{
		Vmlinux: fmt.Sprintf("%s/%s.vmlinux", DefaultKerneldir, name),
		Initrd:  fmt.Sprintf("%s/%s.initrd", DefaultKerneldir, name),
		Config:  fmt.Sprintf("%s/%s.config", DefaultKerneldir, name),
	}
}

func (k *KernelConf) CheckVmlinux() error {
	return checkFile(k.Vmlinux, "vmlinux", ErrKernelExists, nil)
}

func (k *KernelConf) HasInitrd() error {
	return checkFile(k.Initrd, "initrd", nil, ErrInitrdNotFound)
}

func (k *KernelConf) HasVmlinux() error {
	return checkFile(k.Vmlinux, "vmlinux", nil, ErrKernelNotFound)
}
