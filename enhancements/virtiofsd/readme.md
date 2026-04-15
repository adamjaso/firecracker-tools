# Sharing a host directory with Firecracker guest with virtiofsd

These are some notes on using viritofsd with Firecracker.

## TL;DR

Firecracker has not supported host-guest directory sharing, i.e. no `9p` fs sharing.
However, support for `virtiofsd` [^1] is in development [^2] and is currently functional
in a fork by github user `superserve-ai`. Docs are here. [^3]

## Requirements

Guest kernel needs `CONFIG_VIRTIO_FS=y`.

I forked the fork [^4] with virtiofsd support and merged in the main branch of `firecracker`. [^5]

You will need to build either their fork [^4] or my fork [^5] to have a Firecracker binary with `virtiofsd` support.

You will need to install `virtiofsd` from your package manager or build from the source. [^1]

## How to

### Start virtiofsd

Create a directory `./share` to be shared with the guest, and create a test file to observe that sharing works --

```
mkdir -p ./share
echo abc > ./share/test
ls -l ./share/test
/usr/libexec/virtiofsd --socket-path=/tmp/virtiofsd.sock --shared-dir=./share --tag=myfs --log-level=debug
```

Note that a running `virtiofsd` process is only valid for a single VM boot. If the VM connects, and later
is killed or shut down, the `virtiofsd` daemon exits, and must be restarted.

### Configure Firecracker

Add this snippet to the VM configuration file --

```
"vhost-user-devices": [
    {
        "id": "fs0",
        "device_type": 26,
        "socket": "/tmp/virtiofsd.sock",
        "num_queues": 2,
        "queue_size": 256
    }
]
```

The `vhost-user` docs/defaults [^3] need some brushing up. The docs example indicates `num_queues=1`, but in my testing,
I found that only `num_queues=2` works.

I tried `num_queues=1` and the kernel would panic.

```
[    0.415314] virtiofs virtio2: discovered new tag: myfs
[    0.416059] virtiofs virtio2: probe with driver virtiofs failed with error -2
```

I tried `num_queues=3`, and the kernel booted and the tag mounted, but `ls` in the mount point hung indefinitely.

### Start firecracker

I cloned my Firecracker fork-of-a-fork [^5] into `./firecracker-with-virtiofsd` and ran `cargo build` in the clone.
The Firecracker binary is `./firecracker-with-virtiofsd/build/cargo_target/debug/firecracker`.

```
doas ./firecracker-with-virtiofsd/build/cargo_target/debug/firecracker \
    --api-sock /var/run/firecracker/test.sock \
    --config-file /var/lib/firecracker/conf/test.json \
    --id test
```

The `test.json` is as follows --

```
{
  "boot-source": {
    "boot_args": "rootfstype=ext4 rw console=ttyS0",
    "kernel_image_path": "/var/lib/firecracker/kernel/linux-6.19.11.vmlinux"
  },
  "drives": [
    {
      "drive_id": "/var/lib/firecracker/disk/test.img",
      "io_engine": "Sync",
      "is_read_only": false,
      "is_root_device": true,
      "path_on_host": "/var/lib/firecracker/disk/test.img"
    }
  ],
  "logger": {
    "level": "Info",
    "log_path": "/var/log/firecracker/test.log",
    "show_level": true
  },
  "machine-config": {
    "mem_size_mib": 8192,
    "smt": false,
    "vcpu_count": 8
  },
  "network-interfaces": [
    {
      "guest_mac": "00:bb:cc:dd:ee:ff",
      "host_dev_name": "tap1",
      "iface_id": "tap1"
    }
  ],
  "mmds-config": {
    "network_interfaces": ["tap1"],
    "version": "V1"
  },
  "vhost-user-devices": [
    {
        "id": "fs0",
        "device_type": 26,
        "socket": "/tmp/virtiofsd.sock",
        "num_queues": 2,
        "queue_size": 256
    }
  ]
}
```

### Mount the shared directory

Now that you've started `virtiofsd` and `firecracker`, prepare to mount the shared directory.

If all goes well, you should see the kernel detect the `myfs` virtiofs tag.

```
...
[    0.537750] virtiofs virtio2: discovered new tag: myfs
...
Welcome to Alpine Linux 3.23
Kernel 6.19.11 on x86_64 (/dev/ttyS0)

firecracker login: root
Welcome to Alpine!

The Alpine Wiki contains a large amount of how-to guides and general
information about administrating Alpine systems.
See <https://wiki.alpinelinux.org/>.

You can setup the system with the command: setup-alpine

You may change this message by editing /etc/motd.

login[599]: root login on 'ttyS0'
firecracker:~# mount -t virtiofs myfs /mnt
firecracker:~# ls /mnt
test
firecracker:~#
```

[^1]: https://gitlab.com/virtio-fs/virtiofsd
[^2]: https://github.com/firecracker-microvm/firecracker/pull/5773
[^3]: https://github.com/superserve-ai/firecracker/blob/388b31de6d4d80361c19c56fa5fbaaf953300585/docs/vhost-user.md
[^4]: https://github.com/superserve-ai/firecracker
[^5]: https://github.com/adamjaso/firecracker-with-virtiofsd
