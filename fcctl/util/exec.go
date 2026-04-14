package util

import (
	"context"
	"fmt"
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
