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
	cmd.rootfs = fclib.NewRootfs(cmd.nameF)
}

func (cmd *RootfsCommand) Exec(ctx context.Context) {
	showErr(errUnknownCommand)
}

func (cmd *RootfsCommand) Edit() {
	util.ExecCommand(context.Background(), fmt.Sprintf("tar -tzvf %s", cmd.rootfs.Tarball))
}

func (cmd *RootfsCommand) List() {
	mainList("rootfs", fclib.DefaultConfdir, ".rootfs.tar.gz")
}

func (cmd *RootfsCommand) Install() {
	if err := cmd.rootfs.CheckTarball(); err != nil {
		showErr(err)
	}
	installFiles := [][2]string{{cmd.fileF, cmd.rootfs.Tarball}}
	installData(nil, installFiles, nil)
	log.Printf("installed rootfs %q", cmd.nameF)
}
