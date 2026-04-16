package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func StartCommand(ctx context.Context, command string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "sh", "-xec", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, cmd.Start()
}

func ExecCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-xec", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecCompress(ctx context.Context, flags, src, dst string) error {
	// ensure outfile does not exist, to avoid accidental or malicious file overwriting
	if _, err := os.Stat(dst); err == nil {
		return fmt.Errorf("compress output file %s exists, refusing to overwrite", dst)
	} else if !os.IsNotExist(err) {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()
	cmd := exec.CommandContext(ctx, "gzip", flags, src)
	cmd.Stdin = os.Stdin
	cmd.Stdout = out
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ExecCommands(ctx context.Context, commands ...string) error {
	for _, command := range commands {
		if err := ExecCommand(ctx, command); err != nil {
			fmt.Println(command, err)
			return err
		}
	}
	return nil
}

func EditFile(fileF string) {
	cmd := exec.Command(os.Getenv("EDITOR"), fileF)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	AssertNoErr(cmd.Run())
}

func AssertNoErr(err error) {
	if err != nil {
		log.Output(2, fmt.Sprintf("%v\n", err))
		os.Exit(1)
	}
}
