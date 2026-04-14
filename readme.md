# Firecracker Utility

This repository provides tools to run a [Firecracker VM](https://github.com/firecracker-microvm/firecracker).

## TL;DR

- `./fcctl` is a CLI tool that provides some utilities to play with firecracker VMs
- `./firecracker` provides a local build of the firecracker binary
- `./kernel` provides a demo kernel build you can use to boot a firecracker VM and also provides tooling for building your own custom kernel
- `./rootfs` provides a demo of how to build a disk image you can use as the rootfs in a firecracker VM

See the [Makefile](./Makefile) for demo details.

## Dependencies

Depending on your system, you may need to install some of these packages.

- General
    - GNU Make - `make`
    - GCC build toolchain (Ubuntu: `apt install build-essential`, Alpine: `apk add build-base`, etc)

- fcctl
    - requires `go` toolchain

- Firecracker
    - requires Rust compiler and the Rust package manager `cargo`
    - may require `libseccomp-devel` and other C library sources, depending on what you already have installed

- Kernel
    - requires GCC
    - requires `flex`, `bison`, `ncurses-dev` (if you want menuconfig)

- Rootfs
    - requires GNU `tar`
    - requires `wget`

These may not be comprehensive, but consider this a heads up.

## Getting started

You need to run this as root since

- `fcctl` creates directories in `/var/lib/firecracker` and uses `chroot`
- `firecracker` uses KVM

### Prepare

Initialize the demonstration using the `make` target `init-demo`.

```
sudo make init-demo
```

This runs all the `fcctl` commands necessary to create a root disk and a vm configuration

### Start the VM

Start the demonstration using the `make` target `start-demo`.

```
sudo make start-demo
```

If all goes well, you should have a running Firecracker VM, but chances are that you will need to do a bit of troubleshooting.
