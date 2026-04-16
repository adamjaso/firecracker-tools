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
	}
)

func (cmd *KernelCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Installed kernel name")
	flag.StringVar(&cmd.kernelF, "f", "", "Install kernel vmlinux")
	flag.StringVar(&cmd.initrdF, "i", "", "Install initrd/initramfs (optional)")
	flag.StringVar(&cmd.configF, "c", "", "Installed kernel config (optional)")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *KernelCommand) Edit() {
	util.EditFile(fclib.NewKernel(cmd.nameF).Config)
}

func (cmd *KernelCommand) Exec(ctx context.Context) {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *KernelCommand) List() {
	mainList("kernel", fclib.DefaultKerneldir, ".vmlinux")
}

func (cmd *KernelCommand) Install() {
	kernel := fclib.NewKernel(cmd.nameF)
	util.AssertNoErr(kernel.CheckVmlinux())
	installFiles := [][2]string{
		{cmd.kernelF, kernel.Vmlinux},
	}
	if cmd.initrdF != "" {
		installFiles = append(installFiles, [2]string{cmd.initrdF, kernel.Initrd})
	}
	installData(nil, installFiles, nil)
	log.Printf("installed kernel %q", cmd.nameF)
}
