# Building a root disk image for Firecracker

A basic root disk image for a Firecracker VM is constructed of a rootfs, like Alpine's minirootfs
with customizations that install an init system, enable networking, configure the `/dev` mount,
enable boot services, and install packages, etc.

## How it works

Customizations are made by extracting the rootfs to a directory and then chrooting to the directory
and running a script that applies customizations.

This demo provides a `make build` command that

1. downloads the Alpine minirootfs (3.23.3 as of this writing)
2. copies a chroot script to the rootfs extract
3. chroot runs the script to apply customizations
4. the rootfs is tarred and then extracted to an ext4 formatted raw disk image

This ext4 raw disk image is ready to boot as a Firecracker VM root disk.

## init systems

This demo provides a `chroot` demo script using the `openrc` init system, and the `s6` init system.

**NOTE**: the `chroot_s6.sh` script requires passing a kernel cmdline option `init=/init` otherwise
the default Alpine `/sbin/init` will be called and try to boot OpenRC.
