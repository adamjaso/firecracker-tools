package fclib

import (
	"context"
	"fmt"

	"fcctl/util"
)

func ExecIntoChroot(ctx context.Context, buildDir, srcTarball, chrootScript string) error {
	return util.ExecCommands(
		ctx,
		fmt.Sprintf("mkdir -p %s", buildDir),
		fmt.Sprintf("tar -C %s -xzvf %s", buildDir, srcTarball),
		fmt.Sprintf("cp -av %s %s/chroot.sh", chrootScript, buildDir),
		fmt.Sprintf("chroot %s sh", buildDir),
	)
}

func BuildDisk(ctx context.Context, rootfs RootfsConf, chroot ChrootConf, disk DiskConf, diskSize int) error {
	if err := disk.CheckFile(); err != nil {
		return err
	}
	return util.ExecCommands(
		ctx,
		fmt.Sprintf("rmdir %s 2>/dev/null || true", chroot.Dir),
		fmt.Sprintf("mkdir -p %s 2>/dev/null", chroot.Dir),
		fmt.Sprintf("umount %s 2>/dev/null || true", chroot.Dir),
		fmt.Sprintf("dd if=/dev/zero of=%s bs=1M count=%d", disk.File, diskSize),
		fmt.Sprintf("mkfs.ext4 %s", disk.File),
		fmt.Sprintf("mount %s %s", disk.File, chroot.Dir),
		fmt.Sprintf("tar -C %s -xzvf %s", chroot.Dir, rootfs.Tarball),
		fmt.Sprintf("cp -av %s %s/chroot.sh", chroot.Script, chroot.Dir),
		fmt.Sprintf("chroot %s sh -ex chroot.sh", chroot.Dir),
	)
}
