package util

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func ExecCommand(ctx context.Context, command string) error {
	cmd := exec.CommandContext(ctx, "sh", "-xec", command)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
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
