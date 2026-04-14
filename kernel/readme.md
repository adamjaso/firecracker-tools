# Building kernels for Firecracker

The `Makefile` attempts to provide a starting point for building your own kernel for Firecracker.

It downloads a kernel binary, extracts it to a directory, uses the base `linux-6.19.11-config`
as it's `olddefconfig` from which you can run `make menuconfig` to tweak options according to your needs.

The `linux-6.19.11-vmlinux` binary should be a working kernel binary suitable for booting a
Firecracker VM.

Run `make` to attempt to build the kernel from the provided config.
