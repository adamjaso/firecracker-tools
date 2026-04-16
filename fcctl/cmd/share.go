package cmd

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	ShareCommand struct {
		nameF   string
		sockidF string
		cmdF    string
	}
)

func (cmd *ShareCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Share name")
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
	if sockidF := flag.Arg(1); sockidF != "" {
		cmd.sockidF = sockidF
	}
}

func (cmd *ShareCommand) Edit() {
	share := fclib.NewShare(cmd.nameF)
	log.Printf("entering share %q directory %s...", share.Name, share.Dir)
	_ = util.ExecCommand(context.Background(), fmt.Sprintf("cd %s && env PS1=\"[%s] $PS1\" bash", share.Dir, share.Name))
}

func (cmd *ShareCommand) List() {
	mainList("share", fclib.DefaultSharedir, ".share")
}

func (cmd *ShareCommand) Install() {
	share := fclib.NewShare(cmd.nameF)
	util.AssertNoErr(share.CheckDir())
	util.AssertNoErr(os.MkdirAll(share.Dir, 0o755))
	log.Printf("installed share %s", share.Dir)
}

func (cmd *ShareCommand) StartVirtiofs(ctx context.Context, sockidF string) {
	share := fclib.NewShare(cmd.nameF)
	log.Printf("starting virtiofsd for %q...", share.Name)
	util.AssertNoErr(share.Exec(ctx, sockidF, false))
	log.Printf("exited virtiofsd for %q", share.Name)
}

func (cmd *ShareCommand) Exec(ctx context.Context) {
	share := fclib.NewShare(cmd.nameF)
	switch flag.Arg(2) {
	case "clean":
		_, err := share.GetSock(cmd.sockidF, true)
		util.AssertNoErr(err)
	default:
		cmd.StartVirtiofs(ctx, cmd.sockidF)
	}
}
