package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

var DefaultVirtiofsdBin = "/usr/libexec/virtiofsd"

func init() {
	if bin := os.Getenv("VIRTIOFSD_BIN"); bin != "" {
		DefaultVirtiofsdBin = bin
	}
}

func GetVirtiofsd(virtiofsdBin, socket, tag, dirname, additionalFlags string) string {
	return fmt.Sprintf("%s --socket-path=%s --tag=%s --shared-dir=%s %s", virtiofsdBin, socket, tag, dirname, additionalFlags)
}

func StartVirtiofsd(ctx context.Context, virtiofsdBin, socket, tag, dirname, additionalFlags string) (*exec.Cmd, error) {
	if virtiofsdBin == "" {
		virtiofsdBin = DefaultVirtiofsdBin
	}
	return StartCommand(ctx, GetVirtiofsd(virtiofsdBin, socket, tag, dirname, additionalFlags))
}
