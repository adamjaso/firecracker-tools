package cmd

import (
	"context"
	"flag"
	"log"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	ChrootCommand struct {
		nameF   string
		scriptF string
		rootfsF string
		chroot  *fclib.ChrootConf
	}
)

func (cmd *ChrootCommand) Parse() {
	if flag.NArg() == 0 {
		flag.StringVar(&cmd.nameF, "N", "", "Installed chroot name")
		flag.StringVar(&cmd.scriptF, "f", "", "Installed chroot script")
	} else {
		flag.StringVar(&cmd.rootfsF, "rootfs", "", "Rootfs name")
	}
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
	cmd.chroot = fclib.NewChroot(cmd.nameF)
}

func (cmd *ChrootCommand) Edit() {
	util.EditFile(cmd.chroot.Script)
}

func (cmd *ChrootCommand) Exec(ctx context.Context) {
	rootfs := fclib.NewRootfs(cmd.rootfsF)
	util.AssertNoErr(fclib.ExecIntoChroot(ctx, cmd.chroot.Dir, rootfs.Tarball, cmd.chroot.Script))
}

func (cmd *ChrootCommand) List() {
	mainList("chroot", fclib.DefaultConfdir, ".chroot.sh")
}

func (cmd *ChrootCommand) Install() {
	util.AssertNoErr(cmd.chroot.CheckScript())
	installDirs := []string{cmd.chroot.Dir}
	installFiles := [][2]string{{cmd.scriptF, cmd.chroot.Script}}
	installData(installDirs, installFiles, nil)
	log.Printf("installed chroot script %q", cmd.nameF)
}
