# fcctl

This tool manages firecracker VM configuration files as well as some configs
that make it easier to build VM configs.

## Why not firectl?

The [firectl](https://github.com/firecracker-microvm/firectl) tool is a good start for playing with Firecracker,
but I wanted more functionality and persistence for tinkering with Firecracker, and I just wanted to learn
more about the basic `firecracker` flags and Go SDK.

## Getting started

Run `make` to build the binary.

Run the binary `./fcctl -h` and inspect the help flags. Specifying a subcommand provides additional help details, i.e. `./fcctl vm -h`.

## Configs

The configs that `fcctl` manages are as follows.

- Rootfs:
    - config: `/var/lib/firecracker/conf/*.rootfs.tar.gz`
- Chroots:
    - script: `/var/lib/firecracker/conf/*.chroot.sh`]
    - mount: `/var/run/firecracker/*.chroot`
- Disks:
    - raw disk image: `/var/lib/firecracker/disk/*.img`
- Kernels: `/var/lib/firecracker/kernel/*`
    - kernel binary: `/var/lib/firecracker/kernel/*.vmlinux`
    - kernel config: `/var/lib/firecracker/kernel/*.config`
    - optional initrd: `/var/lib/firecracker/kernel/*.initrd`
- Images: `/var/lib/firecracker/image/*`
    - image file:  `/var/lib/firecracker/image/*.img.gz`
- Shares: `/var/lib/firecracker/share/*`
    - share directory:  `/var/lib/firecracker/share/*.share`
- VMs: `/var/lib/conf/firecracker/*.json`
    - vm config: `/var/lib/firecracker/conf/*.json`
    - vm socket: `/var/run/firecracker/*.sock`

### Rootfs

A `rootfs` is a tarball (as of now, it must be `.tar.gz`) that is used to create a base filesystem for a root disk image.
This `rootfs` can be as simple as the [Alpine Linux](https://alpinelinux.org/downloads/) minirootfs build.

### Chroots

A `chroot` is a shell script that executes inside a directory that has been pre-populated with the contents of a `rootfs` tarball.
The `chroot` shell script customizes the rootfs to prepare a root disk for booting in a Firecracker VM.

### Images

An `image` is a compressed snapshot of a `disk`. It is intended to be a base image for creating VMs. For example, when creating a 
new VM, the user can reference an image `-I` which will be used to auto-create a new disk for the VM.

### Disks

A `disk` is a bootable root disk that can be booted by a Firecracker VM. It is an `ext4` formatted raw disk image i.e. `dd if=/dev/zero`
and contains no partitions.

A disk is constructed from a `rootfs` and `chroot` by

1. creating a raw disk image
2. formatting the disk image
3. mounting the disk image
4. extracting a `rootfs` into the mount
5. running a `chroot` script in the mount with `chroot` command
6. unmount the disk image

A disk can also be created by decompressing an `image`.

### Kernels

A `kernel` is a bootable kernel that has enabled all `CONFIG_VIRTIO*` related features. Firecracker relies mainly on `virtio` drivers
and so these and all other required kernel functionality i.e. nftables, ext4, etc should be compiled into the kernel i.e. all config options `=y`.
A [working x86_64 kernel](./linux-6.19.11-vmlinux) has been provided along with its [source config](./linux-6.19.11-config).

You can use this demo kernel as a base for booting test VMs.

You can use this demo config as a base for tweaking further kernel features according to your needs.

### VMs

A `vm` is a valid [Firecracker VM configuration file](https://pkg.go.dev/github.com/firecracker-microvm/firecracker-go-sdk/client/models#FullVMConfiguration) that may be passed to the `firecracker --config-file` flag.

A `vm` is constructed from a `kernel` and a `disk`. The `vm` command provides a `start` command that will construct
and execute a `firecracker` command that runs the VM.

The `vm` default kernel `cmdline` should specify `console=ttyS0` i.e. serial console since that Firecracker stdin
connects to the VM serial port.

The `vm` command provides a command to stop the VM using the Firecracker unix socket.

The `vm` command provides a command to curl the VM Firecracker unix socket with a custom HTTP method, URL path, and body.

Valid HTTP parameters can be found in the [Firecracker Swagger docs](https://github.com/firecracker-microvm/firecracker-go-sdk/blob/main/client/swagger.yaml#L30).

### Shares

A `share` is a shared directory with the Firecracker VM. This is currently only possible using a forked version of Firecracker [^1]
which has added support for `virtiofsd`. `fcctl` supports starting the `virtiofsd` daemon to enable host/guest directory sharing.
The forked version of Firecracker supporting this feature is currently pending review.

[^1]: [Firecracker fork supporting virtiofsd](https://github.com/firecracker-microvm/firecracker/pull/5773)
