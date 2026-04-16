package cmd

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"fcctl/fclib"
	"fcctl/util"

	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type (
	VmCommand struct {
		nameF    string
		kernelF  string
		cmdlineF string
		diskF    string
		shareF   string
		vcpuF    int64
		memF     int64
		tapF     string
		metricsF string
		mmdsF    bool
		opts     util.VmOpts

		cmdnameF string
		argsF    []string
	}
)

func (cmd *VmCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Vm name")
	flag.StringVar(&cmd.kernelF, "K", "", "Kernel name")
	flag.StringVar(&cmd.diskF, "D", "", "Disk name")
	flag.StringVar(&cmd.shareF, "S", "", "(optional) Include virtiofs shared directory config (virtiofsd required)")
	flag.StringVar(&cmd.cmdlineF, "cmdline", "rootfstype=ext4 rw console=ttyS0", "(optional) Kernel cmdline")
	flag.Int64Var(&cmd.vcpuF, "n", 1, "(optional) Number of vCPU")
	flag.Int64Var(&cmd.memF, "m", 512, "(optional) VM memory in MiB")
	flag.StringVar(&cmd.tapF, "tap-device", "tap0/aa:bb:cc:dd:ee:ff", "(optional) Tap interface name")
	flag.StringVar(&cmd.metricsF, "metrics", "", "(optional) Metrics file path")
	flag.BoolVar(&cmd.mmdsF, "mmds", false, "(optional) Include MMDS config. Only valid with -tap-device")
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
}

func (cmd *VmCommand) Exec(ctx context.Context) {
	if len(flag.Args()) < 1 {
		flag.Usage()
	}
	vm := fclib.New(cmd.nameF)
	if cmd.cmdnameF == "start" {
		_ = util.RunFirecrackerCommand(ctx, vm.Sock, "stop")
		util.AssertNoErr(util.WaitUntilState(ctx, vm.Sock, models.InstanceInfoStateNotStarted))
		// try to cleanup
		util.AssertNoErr(vm.Clean())
	}
	if err := vm.CheckSock(); err != nil {
		util.AssertNoErr(util.RunFirecrackerCommand(ctx, vm.Sock, cmd.cmdnameF, cmd.argsF...))
		return
	}
	// no socket, vm is probably not running

	wgShareExit := sync.WaitGroup{} // wait for virtiofsd and vm to exit
	shares := vm.GetShares()
	shareCmds := make([]*exec.Cmd, len(shares))
	if len(shares) > 0 {
		for i, share := range shares {
			log.Printf("starting virtiofsd for share %q...", share.Name)
			cmd, err := share.Start(ctx, vm.Name, false)
			util.AssertNoErr(err)
			shareCmds[i] = cmd
		}
		wgShareExit.Add(1)
		go func() {
			for i, cmd := range shareCmds {
				share := shares[i]
				if err := cmd.Wait(); err != nil {
					log.Printf("exited virtiofsd for share %q with %d", share.Name, cmd.ProcessState.ExitCode())
				} else {
					log.Printf("exited virtiofsd for share %q", share.Name)
				}
			}
			wgShareExit.Done()
		}()
	}
	log.Printf("starting firecracker for vm %q from %q\n", vm.Name, vm.File)
	builder := vm.GetVMCommandBuilder(fclib.FirecrackerBin)
	vmCmd := builder.Build(ctx)
	log.Printf("firecracker command:\n\t%s\n", strings.Join(vmCmd.Args, " \\\n\t\t"))
	if err := vmCmd.Run(); err != nil {
		log.Printf("exited vm %q firecracker with %d", vm.Name, vmCmd.ProcessState.ExitCode())
	} else {
		log.Printf("exited vm %q firecracker", vm.Name)
	}
	for i, vmCmd := range shareCmds {
		log.Printf("stopping virtiofsd for share %q", shares[i].Name)
		_ = vmCmd.Process.Signal(syscall.SIGTERM)
	}
	wgShareExit.Wait()
}

func (cmd *VmCommand) List() {
	mainList("vm", fclib.DefaultConfdir, ".json")
}

func (cmd *VmCommand) Edit() {
	util.EditFile(fclib.New(cmd.nameF).File)
}

func (cmd *VmCommand) Install() {
	if cmd.kernelF == "" || cmd.diskF == "" {
		flag.Usage()
	}

	vm := fclib.New(cmd.nameF)
	if err := vm.CheckFile(); err != nil {
		util.AssertNoErr(err)
	}
	kernel := fclib.NewKernel(cmd.kernelF)
	util.AssertNoErr(kernel.HasVmlinux())
	opts := util.VmOpts{
		Vcpu:        cmd.vcpuF,
		Mem:         cmd.memF,
		TapDevice:   cmd.tapF,
		MetricsPath: cmd.metricsF,
		EnableMMDS:  cmd.mmdsF,
		LogPath:     vm.Log,
		DiskPath:    fclib.NewDisk(cmd.diskF).File,
		Cmdline:     cmd.cmdlineF,
		KernelPath:  kernel.Vmlinux,
	}
	if err := kernel.HasInitrd(); err == nil {
		opts.InitrdPath = kernel.Initrd
	} else if !errors.Is(err, fclib.ErrInitrdNotFound) {
		util.AssertNoErr(err)
	}

	if cmd.shareF != "" {
		log.Printf("WARNING! virtiofsd support is not yet a mainline firecracker feature. You will need to build Firecracker yourself from a fork that supports virtiofsd.")
		share := fclib.NewShare(cmd.shareF)
		opts.VirtiofsID = cmd.shareF
		opts.VirtiofsSock, _ = share.GetSock(cmd.nameF, false)
	}

	if vmConf, err := util.BuildVm(opts); err != nil {
		util.AssertNoErr(err)
	} else if err := vm.WriteVm(vmConf); err != nil {
		util.AssertNoErr(err)
	}
	log.Printf("wrote config to %s", vm.File)
}
