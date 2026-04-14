package cmd

import (
	"context"
	"flag"
	"log"

	"fcctl/fclib"
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
	showErr(errUnknownCommand)
}

func (cmd *DiskCommand) Exec(ctx context.Context) {
	showErr(errUnknownCommand)
}

func (cmd *DiskCommand) List() {
	mainList("disk", fclib.DefaultDiskdir, ".img")
}

func (cmd *DiskCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Disk name")
	flag.StringVar(&cmd.chrootF, "c", "", "Chroot script name")
	flag.StringVar(&cmd.rootfsF, "r", "", "Rootfs tarball name")
	flag.IntVar(&cmd.sizeF, "s", 4096, "Disk size in MB")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *DiskCommand) Install() {
	d := fclib.NewDisk(cmd.nameF)
	if err := d.CheckFile(); err != nil {
		showErr(err)
	}
	rfs := fclib.NewRootfs(cmd.rootfsF)
	chr := fclib.NewChroot(cmd.chrootF)
	ctx := context.Background()
	if err := fclib.BuildDisk(ctx, *rfs, *chr, *d, cmd.sizeF); err != nil {
		showErr(err)
	}
	log.Printf("installed disk %q", cmd.nameF)
}
