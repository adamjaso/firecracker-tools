package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
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

func showErr(err error) {
	log.Printf("%v\n", err)
	os.Exit(1)
}

func mainList(action, dir, suffix string) {
	files, err := filepath.Glob(fmt.Sprintf("%s/*%s", dir, suffix))
	if err != nil {
		showErr(err)
	}
	for _, fn := range files {
		name := strings.ReplaceAll(filepath.Base(fn), suffix, "")
		fmt.Printf("%s\t%s\n", name, fn)
	}
}

func installData(installDirs []string, installFiles, installContents [][2]string) {
	for _, dir := range installDirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			showErr(err)
		}
	}
	for _, srcdst := range installFiles {
		if err := util.InstallArtifact(srcdst[0], srcdst[1]); err != nil {
			showErr(err)
		}
	}
	for _, fileContents := range installContents {
		if err := os.WriteFile(fileContents[0], []byte(fileContents[1]), 0o640); err != nil {
			showErr(err)
		}
	}
}

func editFile(fileF string) {
	cmd := exec.Command(os.Getenv("EDITOR"), fileF)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		showErr(err)
	}
}
