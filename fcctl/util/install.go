package util

import "os"

func InstallArtifact(srcpath, dstpath string) error {
	if contents, err := os.ReadFile(srcpath); err != nil {
		return err
	} else if err := os.WriteFile(dstpath, contents, 0o640); err != nil {
		return err
	}
	return nil
}
