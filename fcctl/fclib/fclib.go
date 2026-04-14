package fclib

import (
	"os"
	"path/filepath"
)

var (
	Progname       string = filepath.Base(os.Args[0])
	FirecrackerBin string
)

func init() {
	if FirecrackerBin = os.Getenv("FC_BIN"); FirecrackerBin == "" {
		FirecrackerBin = "firecracker"
	}
}
