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
	RootfsCommand struct {
		nameF string
		fileF string

		rootfs *fclib.RootfsConf
	}
)

func (cmd *RootfsCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Rootfs name")
	flag.StringVar(&cmd.fileF, "f", "", "Rootfs source tarball")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
}

func (cmd *RootfsCommand) Exec(ctx context.Context) {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *RootfsCommand) Edit() {
	rootfs := fclib.NewRootfs(cmd.nameF)
	util.ExecCommand(context.Background(), fmt.Sprintf("tar -tzvf %s", rootfs.Tarball))
}

func (cmd *RootfsCommand) List() {
	mainList("rootfs", fclib.DefaultConfdir, ".rootfs.tar.gz")
}

func (cmd *RootfsCommand) Install() {
	rootfs := fclib.NewRootfs(cmd.nameF)
	util.AssertNoErr(rootfs.CheckTarball())
	installFiles := [][2]string{{cmd.fileF, rootfs.Tarball}}
	installData(nil, installFiles, nil)
	log.Printf("installed rootfs %q", cmd.nameF)
}
