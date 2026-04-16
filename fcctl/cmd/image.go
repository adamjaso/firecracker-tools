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
	ImageCommand struct {
		nameF    string
		diskF    string
		cmdnameF string
		cmdargF  string
	}
)

func (cmd *ImageCommand) Parse() {
	flag.StringVar(&cmd.nameF, "N", "", "Image name")
	flag.StringVar(&cmd.diskF, "D", "", "Disk name to compress to an image")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s image [CREATE_FLAGS|NAME COMMAND [ARGS...]]

       %[1]s image - lists existing images

       %[1]s image CREATE_FLAGS - compresses a new image the specified disk

       %[1]s image NAME COMMAND [ARGS] - executes one of the following commands

COMMAND

  rename NEWNAME - renames the image

CREATE_FLAGS

`, fclib.Progname)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if flag.NArg() > 0 && flag.NArg() < 3 {
		flag.Usage()
	}
	if name := flag.Arg(0); name != "" {
		cmd.nameF = name
	}
	if cmdname := flag.Arg(1); cmdname != "" {
		cmd.cmdnameF = cmdname
		cmd.cmdargF = flag.Arg(2)
	}
}

func (cmd *ImageCommand) List() {
	mainList("image", fclib.DefaultImagedir, ".img.gz")
}

func (cmd *ImageCommand) Edit() {
	util.AssertNoErr(errUnknownCommand)
}

func (cmd *ImageCommand) Install() {
	image := fclib.NewImage(cmd.nameF)
	util.AssertNoErr(image.CheckFile())
	disk := fclib.NewDisk(cmd.diskF)
	util.AssertNoErr(disk.Exists())
	util.ExecCompress(context.Background(), "-c", disk.File, image.File)
	log.Printf("compressed image %q from disk %q", cmd.nameF, cmd.diskF)
}

func (cmd *ImageCommand) Exec(ctx context.Context) {
	if cmd.cmdnameF == "" || cmd.cmdargF == "" {
		flag.Usage()
	}
	img := fclib.NewImage(cmd.nameF)
	switch cmd.cmdnameF {
	case "rename":
		newimg := fclib.NewImage(cmd.cmdargF)
		util.AssertNoErr(os.Rename(img.File, newimg.File))
	default:
		flag.Usage()
	}
}
