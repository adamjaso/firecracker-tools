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

		share *fclib.ShareConf
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
	cmd.share = fclib.NewShare(cmd.nameF)
}

func (cmd *ShareCommand) Edit() {
	log.Printf("entering share %q directory %s...", cmd.share.Name, cmd.share.Dir)
	_ = util.ExecCommand(context.Background(), fmt.Sprintf("cd %s && env PS1=\"[%s] $PS1\" bash", cmd.share.Dir, cmd.share.Name))
}

func (cmd *ShareCommand) List() {
	mainList("share", fclib.DefaultSharedir, ".share")
}

func (cmd *ShareCommand) Install() {
	util.AssertNoErr(cmd.share.CheckDir())
	util.AssertNoErr(os.MkdirAll(cmd.share.Dir, 0o755))
	log.Printf("installed share %s", cmd.share.Dir)
}

func (cmd *ShareCommand) StartVirtiofs(ctx context.Context, sockidF string) {
	log.Printf("starting virtiofsd for %q...", cmd.share.Name)
	util.AssertNoErr(cmd.share.Exec(ctx, sockidF, false))
	log.Printf("exited virtiofsd for %q", cmd.share.Name)
}

func (cmd *ShareCommand) Exec(ctx context.Context) {
	switch flag.Arg(2) {
	case "clean":
		_, err := cmd.share.GetSock(cmd.sockidF, true)
		util.AssertNoErr(err)
	default:
		cmd.StartVirtiofs(ctx, cmd.sockidF)
	}
}
