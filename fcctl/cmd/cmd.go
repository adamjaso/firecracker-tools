package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fcctl/util"
)

var errUnknownCommand = errors.New("unknown command")

type (
	Command interface {
		Parse()
		Exec(ctx context.Context)
		List()
		Edit()
		Install()
	}
)

func mainList(action, dir, suffix string) {
	files, err := filepath.Glob(fmt.Sprintf("%s/*%s", dir, suffix))
	util.AssertNoErr(err)
	for _, fn := range files {
		name := strings.ReplaceAll(filepath.Base(fn), suffix, "")
		var size int64
		if stat, err := os.Stat(fn); err == nil {
			size = stat.Size()
		}
		fmt.Printf("%s\t%d\t%s\n", name, size, fn)
	}
}

func installData(installDirs []string, installFiles, installContents [][2]string) {
	for _, dir := range installDirs {
		util.AssertNoErr(os.MkdirAll(dir, 0o755))
	}
	for _, srcdst := range installFiles {
		util.AssertNoErr(util.InstallArtifact(srcdst[0], srcdst[1]))
	}
	for _, fileContents := range installContents {
		util.AssertNoErr(os.WriteFile(fileContents[0], []byte(fileContents[1]), 0o644))
	}
}
