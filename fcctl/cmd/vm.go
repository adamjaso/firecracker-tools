package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"fcctl/fclib"
	"fcctl/util"

	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type (
	VmCommand struct {
		nameF   string
		rootF   string
		kernelF string
		chrootF string
		rootfsF string
		diskF   string
		sizeF   int
		opts    util.VmOpts

		cmdnameF string
		argsF    []string
		vm       *fclib.Conf
	}
)

func (cmd *VmCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Vm name")
	flag.StringVar(&cmd.kernelF, "K", "", "Kernel name prefix")
	flag.StringVar(&cmd.chrootF, "C", "", "Chroot script name")
	flag.StringVar(&cmd.rootfsF, "R", "", "Rootfs tarball name")
	flag.StringVar(&cmd.diskF, "D", "", "Disk name")
	flag.IntVar(&cmd.sizeF, "s", 4096, "Disk size")
	flag.StringVar(&cmd.rootF, "root", "", "Custom disk file path (does not install to disks config dir)")
	flag.Int64Var(&cmd.opts.Vcpu, "smp", 1, "Number of CPU")
	flag.Int64Var(&cmd.opts.Mem, "m", 512, "VM memory in MiB")
	flag.StringVar(&cmd.opts.TapDevice, "tap-device", "tap0/aa:bb:cc:dd:ee:ff", "Tap interface name")
	flag.StringVar(&cmd.opts.KernelPath, "kernel", "./vmlinux", "Kernel vmlinux file path")
	flag.StringVar(&cmd.opts.InitrdPath, "initrd", "", "Kernel initrd/initramfs file path")
	flag.StringVar(&cmd.opts.Cmdline, "append", "rootfstype=ext4 rw console=ttyS0", "Kernel args to append")
	flag.StringVar(&cmd.opts.MetricsPath, "metrics", "", "Metrics file path")
	flag.BoolVar(&cmd.opts.EnableMMDS, "mmds", false, "Include MMDS config. Only valid with -tap-device")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s vm [CREATE_FLAGS|NAME [COMMAND [ARGS...]]]

       %[1]s vm - lists existing configs

       %[1]s vm CREATE_FLAGS - creates a new vm config from the CREATE_FLAGS

       %[1]s vm NAME - opens the config for editing

       %[1]s vm NAME COMMAND - executes one of the following commands

COMMAND
  start - starts the vm
  stop - stops the vm
  status - shows vm status
  curl METHOD PATH - sends an HTTP request to the firecracker unix socket

    METHOD is any valid HTTP method
      PUT, POST, PATCH requests read HTTP body from stdin
    PATH is any valid firecracker URL path

CREATE_FLAGS

`, fclib.Progname)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
	if flag.NArg() > 1 {
		cmd.cmdnameF = flag.Arg(1)
		cmd.argsF = flag.Args()[1:]
	}
	cmd.vm = fclib.New(cmd.nameF)
}

func (cmd *VmCommand) Exec(ctx context.Context) {
	flag.Usage = showRunHelp
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
	}
	if cmd.cmdnameF == "start" {
		_ = util.RunCommand(ctx, cmd.vm.Sock, "stop")
		if err := util.WaitUntilState(ctx, cmd.vm.Sock, models.InstanceInfoStateNotStarted); err != nil {
			showErr(err)
		}
		// try to cleanup
		if err := cmd.vm.Clean(); err != nil {
			showErr(err)
		}
	}
	if err := cmd.vm.CheckSock(); err != nil {
		if err := util.RunCommand(ctx, cmd.vm.Sock, cmd.cmdnameF, cmd.argsF...); err != nil {
			showErr(err)
		}
		return
	}
	// no socket, vm is probably not running

	log.Printf("starting vm %q from %q\n", cmd.vm.Name, cmd.vm.File)
	builder := cmd.vm.GetVMCommandBuilder(fclib.FirecrackerBin)
	b := builder.Build(ctx)
	log.Printf("starting firecracker:\n\t%s\n", strings.Join(b.Args, " \\\n\t\t"))
	if err := b.Run(); err != nil {
		showErr(err)
	}
}

func (cmd *VmCommand) List() {
	mainList("vm", fclib.DefaultConfdir, ".json")
}

func (cmd *VmCommand) Edit() {
	editFile(cmd.vm.File)
}

func (cmd *VmCommand) Install() {
	cmd.opts.LogPath = cmd.vm.Log
	if cmd.kernelF != "" {
		kernel := fclib.NewKernel(cmd.kernelF)
		if err := kernel.CheckVmlinux(); err != nil {
			showErr(err)
		}
		cmd.opts.KernelPath = kernel.Vmlinux
		if err := kernel.CheckInitrd(); err == nil {
			cmd.opts.InitrdPath = kernel.Initrd
		} else if !errors.Is(err, fclib.ErrInitrdNotFound) {
			showErr(err)
		}
	}
	if cmd.chrootF != "" && cmd.rootfsF != "" && cmd.diskF != "" {
		chroot := fclib.NewChroot(cmd.chrootF)
		rootfs := fclib.NewRootfs(cmd.rootfsF)
		disk := fclib.NewDisk(cmd.diskF)
		ctx := context.Background()
		if err := fclib.BuildDisk(ctx, *rootfs, *chroot, *disk, cmd.sizeF); err != nil {
			showErr(err)
		}
	} else if cmd.diskF != "" {
		cmd.opts.DiskPath = fclib.NewDisk(cmd.diskF).File
	} else if cmd.rootF != "" {
		cmd.opts.DiskPath = cmd.rootF
	} else {
		showErr(errors.New("disk not found"))
	}

	if err := cmd.vm.CheckFile(); err != nil {
		showErr(err)
	} else if vm, err := util.BuildVm(cmd.opts); err != nil {
		showErr(err)
	} else if err := cmd.vm.WriteVm(vm); err != nil {
		showErr(err)
	}
	log.Printf("wrote config to %s\n", cmd.vm.File)
}

func showRunHelp() {
	fmt.Fprintf(os.Stderr, "usage: %s COMMAND NAME [ARGS...]\n\nCOMMAND\n\tstart NAME\n\tstop NAME\n\tstatus NAME\n\tcurl NAME METHOD PATH\n\n", fclib.Progname)
	flag.PrintDefaults()
	os.Exit(1)
}

func StartVm(ctx context.Context, cmdF string) {
	var (
		nameF string
		argsF []string
	)
	flag.Usage = showRunHelp
	flag.Parse()
	if len(flag.Args()) < 1 {
		flag.Usage()
	}
	nameF = flag.Arg(0)
	argsF = flag.Args()[1:]
	c := fclib.New(nameF)

	if cmdF == "start" {
		_ = util.RunCommand(ctx, c.Sock, "stop")
		if err := util.WaitUntilState(ctx, c.Sock, models.InstanceInfoStateNotStarted); err != nil {
			showErr(err)
		}
		// try to cleanup
		if err := c.Clean(); err != nil {
			showErr(err)
		}
	}
	if err := c.CheckSock(); err != nil {
		if err := util.RunCommand(ctx, c.Sock, cmdF, argsF...); err != nil {
			showErr(err)
		}
		return
	}
	// no socket, vm is probably not running

	log.Printf("starting vm %q from %q\n", c.Name, c.File)
	builder := c.GetVMCommandBuilder(fclib.FirecrackerBin)
	cmd := builder.Build(ctx)
	log.Printf("starting firecracker:\n\t%s\n", strings.Join(cmd.Args, " \\\n\t\t"))
	if err := cmd.Run(); err != nil {
		showErr(err)
	}
}
