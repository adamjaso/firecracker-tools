package cmd

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"fcctl/fclib"
	"fcctl/util"
)

type (
	NetCommand struct {
		nameF    string
		subnetF  string
		prefixF  string
		cmdnameF string
	}
)

func (n *NetCommand) Parse() {
	flag.StringVar(&n.nameF, "N", "", "Network name")
	flag.StringVar(&n.subnetF, "subnet", "", "Network subnet")
	flag.StringVar(&n.prefixF, "prefix", "ff:aa:bb:cc:dd", "Network mac prefix")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: %s net [CREATE_FLAGS|NAME COMMAND [ARGS...]]

       %[1]s net - lists existing nets

       %[1]s net CREATE_FLAGS - compresses a new net the specified disk

       %[1]s net NAME COMMAND [ARGS] - executes one of the following commands

COMMAND

  setup - configures the network bridge and taps

  dump - shows the configuration of the bridge and taps

CREATE_FLAGS

`, fclib.Progname)
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	if name := flag.Arg(0); name != "" {
		n.nameF = name
	}
	if cmdname := flag.Arg(1); cmdname != "" {
		n.cmdnameF = cmdname
	}
}

func (n *NetCommand) Exec(ctx context.Context) {
	net := fclib.NewNet(n.nameF)
	util.AssertNoErr(net.Read())
	switch n.cmdnameF {
	case "setup":
		util.AssertNoErr(net.SetupBridge())
	case "dump":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		util.AssertNoErr(enc.Encode(net.GetMembers()))
	}
}

func (n *NetCommand) List() {
	mainList("net", fclib.DefaultNetdir, ".json")
}

func (n *NetCommand) Edit() {
	util.EditFile(fclib.NewNet(n.nameF).File)
}

func (n *NetCommand) Install() {
	net := fclib.NewNet(n.nameF)
	if n.subnetF == "" || n.prefixF == "" {
		flag.Usage()
	}
	util.AssertNoErr(net.CheckFile())
	net.Subnet = n.subnetF
	net.GuestMacPrefix = n.prefixF
	util.AssertNoErr(net.Write())
}
