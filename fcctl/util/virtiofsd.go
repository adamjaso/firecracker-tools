package util

import (
	"context"
	"fmt"
	"os"
)

var DefaultVirtiofsdBin = "/usr/libexec/virtiofsd"

func init() {
	if bin := os.Getenv("VIRTIOFSD_BIN"); bin != "" {
		DefaultVirtiofsdBin = bin
	}
}

func RunVirtiofsd(ctx context.Context, virtiofsdBin, socket, tag, dirname, additionalFlags string) error {
	if virtiofsdBin == "" {
		virtiofsdBin = DefaultVirtiofsdBin
	}
	return ExecCommand(ctx, fmt.Sprintf("%s --socket-path=%s --tag=%s --shared-dir=%s %s", virtiofsdBin, socket, tag, dirname, additionalFlags))
}
