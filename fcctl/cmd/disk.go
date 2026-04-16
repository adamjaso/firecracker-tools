package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	DiskCommand struct {
		nameF   string
		chrootF string
		rootfsF string
		imageF  string
		sizeF   int
	}
)

func (cmd *DiskCommand) Edit() {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *DiskCommand) Exec(ctx context.Context) {
	disk := fclib.NewDisk(cmd.nameF)
	mountAndShell := fmt.Sprintf(`mount %[2]s %[3]s && cd %[3]s && env PS1="[disk %[1]s] $PS1" bash`, disk.Name, disk.File, "/mnt")
	util.AssertNoErr(util.ExecCommand(context.Background(), mountAndShell))
}

func (cmd *DiskCommand) List() {
	mainList("disk", fclib.DefaultDiskdir, ".img")
}

func (cmd *DiskCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Disk name")
	flag.StringVar(&cmd.chrootF, "C", "", "Chroot script name")
	flag.StringVar(&cmd.rootfsF, "R", "", "Rootfs tarball name")
	flag.StringVar(&cmd.imageF, "I", "", "Image name")
	flag.IntVar(&cmd.sizeF, "s", 4096, "Disk size in MB")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *DiskCommand) Install() {
	disk := fclib.NewDisk(cmd.nameF)
	util.AssertNoErr(disk.CheckFile())
	ctx := context.Background()
	if cmd.imageF != "" {
		img := fclib.NewImage(cmd.imageF)
		log.Printf("creating from image %q", img.Name)
		util.AssertNoErr(disk.InstallFromImage(ctx, img))
	} else {
		rfs := fclib.NewRootfs(cmd.rootfsF)
		chr := fclib.NewChroot(cmd.chrootF)
		log.Printf("creating from rootfs %q with chroot %q", rfs.Name, chr.Name)
		util.AssertNoErr(fclib.BuildDisk(ctx, *rfs, *chr, *disk, cmd.sizeF))
	}
	log.Printf("installed disk %q", cmd.nameF)
}
