package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"fcctl/cmd"
	"fcctl/fclib"
)

func showHelp() {
	fmt.Fprintf(os.Stderr, "usage: %s vm|kernel|chroot|rootfs|disk [flags]\n", fclib.Progname)
	os.Exit(1)
}

func showManageHelp(action string) func() {
	return func() {
		fmt.Fprintf(os.Stderr, `usage: %s %s [CREATE_FLAGS|NAME]]

       %[1]s %[2]s - lists existing configs

       %[1]s %[2]s CREATE_FLAGS - creates a new %[2]s config from the CREATE_FLAGS

       %[1]s %[2]s NAME - opens the %[2]s config for editing, if editing is supported

CREATE_FLAGS

`, fclib.Progname, action)
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Flags() | log.Lshortfile)
	if len(os.Args) < 2 {
		showHelp()
	}
	_ = fclib.InitConfdirs()
	fclib.Progname = os.Args[0]
	action := os.Args[1]
	os.Args = os.Args[1:]
	var c cmd.Command
	switch action {
	case "vm":
		c = &cmd.VmCommand{}
	case "kernel":
		c = &cmd.KernelCommand{}
	case "chroot":
		c = &cmd.ChrootCommand{}
	case "rootfs":
		c = &cmd.RootfsCommand{}
	case "disk":
		c = &cmd.DiskCommand{}
	default:
		showHelp()
	}
	flag.Usage = showManageHelp(action)
	if len(os.Args) == 1 {
		c.List()
		return
	}
	c.Parse()
	if flag.NArg() == 1 {
		c.Edit()
	} else if flag.NArg() == 2 {
		ctx := context.Background()
		c.Exec(ctx)
	} else {
		c.Install()
	}
}
