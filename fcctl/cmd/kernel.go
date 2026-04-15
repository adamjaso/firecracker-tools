package cmd

import (
	"context"
	"flag"
	"log"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	KernelCommand struct {
		nameF   string
		kernelF string
		initrdF string
		configF string

		kernel *fclib.KernelConf
	}
)

func (cmd *KernelCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Installed kernel name")
	flag.StringVar(&cmd.kernelF, "kernel", "", "Install kernel vmlinux")
	flag.StringVar(&cmd.initrdF, "initrd", "", "Install initrd/initramfs (optional)")
	flag.StringVar(&cmd.configF, "config", "", "Installed kernel config (optional)")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
	cmd.kernel = fclib.NewKernel(cmd.nameF)
}

func (cmd *KernelCommand) Edit() {
	util.EditFile(cmd.kernel.Config)
}

func (cmd *KernelCommand) Exec(ctx context.Context) {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *KernelCommand) List() {
	mainList("kernel", fclib.DefaultKerneldir, ".vmlinux")
}

func (cmd *KernelCommand) Install() {
	installFiles := [][2]string{
		{cmd.kernelF, cmd.kernel.Vmlinux},
	}
	if cmd.initrdF != "" {
		installFiles = append(installFiles, [2]string{cmd.initrdF, cmd.kernel.Initrd})
	}
	installData(nil, installFiles, nil)
	log.Printf("installed kernel %q", cmd.nameF)
}
