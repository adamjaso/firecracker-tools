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
	}
)

func (cmd *ChrootCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Installed chroot name")
	flag.StringVar(&cmd.scriptF, "f", "", "Installed chroot script")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *ChrootCommand) Edit() {
	util.EditFile(fclib.NewChroot(cmd.nameF).Script)
}

func (cmd *ChrootCommand) Exec(ctx context.Context) {
	chroot := fclib.NewChroot(cmd.nameF)
	rootfs := fclib.NewRootfs(cmd.rootfsF)
	util.AssertNoErr(fclib.ExecIntoChroot(ctx, chroot.Dir, rootfs.Tarball, chroot.Script))
}

func (cmd *ChrootCommand) List() {
	mainList("chroot", fclib.DefaultConfdir, ".chroot.sh")
}

func (cmd *ChrootCommand) Install() {
	chroot := fclib.NewChroot(cmd.nameF)
	util.AssertNoErr(chroot.CheckScript())
	installDirs := []string{chroot.Dir}
	installFiles := [][2]string{{cmd.scriptF, chroot.Script}}
	installData(installDirs, installFiles, nil)
	log.Printf("installed chroot script %q", cmd.nameF)
}
