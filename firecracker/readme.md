# Build firecracker binary

This just builds the Firecracker binary locally from the source.

Run `make` to build the binary.

The Firecracker source is cloned to `./firecracker`.

The binary is output to `./firecracker/build/cargo_target/debug/firecracker`.

## virtiofsd support

You can also build Firecracker with virtiofsd support using `make ENABLE_VIRTIOFS=1`.
See the [virtiofsd notes](../enhancements/virtiofsd/).
