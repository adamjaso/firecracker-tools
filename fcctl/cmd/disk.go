package cmd

import (
	"context"
	"flag"
	"log"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	DiskCommand struct {
		nameF   string
		chrootF string
		rootfsF string
		sizeF   int
	}
)

func (cmd *DiskCommand) Edit() {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *DiskCommand) Exec(ctx context.Context) {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *DiskCommand) List() {
	mainList("disk", fclib.DefaultDiskdir, ".img")
}

func (cmd *DiskCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Disk name")
	flag.StringVar(&cmd.chrootF, "C", "", "Chroot script name")
	flag.StringVar(&cmd.rootfsF, "R", "", "Rootfs tarball name")
	flag.IntVar(&cmd.sizeF, "s", 4096, "Disk size in MB")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *DiskCommand) Install() {
	disk := fclib.NewDisk(cmd.nameF)
	util.AssertNoErr(disk.CheckFile())
	rfs := fclib.NewRootfs(cmd.rootfsF)
	chr := fclib.NewChroot(cmd.chrootF)
	ctx := context.Background()
	util.AssertNoErr(fclib.BuildDisk(ctx, *rfs, *chr, *disk, cmd.sizeF))
	log.Printf("installed disk %q", cmd.nameF)
}
