# Screen dump of using fcctl to build a VM

This is a screen dump of using the [fcctl](../../fcctl/) tool to create a Firecracker VM.

This is on an Alpine linux host, and assumes Go 1.26 is installed and in the path.

Install packages

```
alpine-hypervisor:~/firecracker-tools# apk add git cargo rust make build-base libseccomp-dev bash clang-libclang curl tar virtiofsd
```

Build fcctl

```
alpine-hypervisor:~/firecracker-tools# make -C fcctl build
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl
usage: fcctl rootfs|chroot|kernel|disk|share|vm [FLAGS]

```

Download Alpine minirootfs

```
alpine-hypervisor:~/firecracker-tools# wget https://dl-cdn.alpinelinux.org/alpine/v3.23/releases/x86_64/alpine-minirootfs-3.23.4-x86_64.tar.gz
Connecting to dl-cdn.alpinelinux.org (151.101.194.132:443)
saving to 'alpine-minirootfs-3.23.4-x86_64.tar.gz'
alpine-minirootfs-3. 100% |***************************************************************************************************************************************************************************************************************| 3628k  0:00:00 ETA
'alpine-minirootfs-3.23.4-x86_64.tar.gz' saved
```

Register the rootfs

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl rootfs -N alpine -f alpine-minirootfs-3.23.4-x86_64.tar.gz
2026/04/16 00:04:31 rootfs.go:48: installed rootfs "alpine"
```

Register the Alpine chroot script

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl chroot -N openrc -f rootfs/chroot_openrc.sh
2026/04/16 00:04:52 chroot.go:53: installed chroot script "openrc"
```

Register the kernel binary

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl kernel -N linux-6.19.11 -config kernel/linux-6.19.11-config -f kernel/linux-6.19.11-vmlinux
2026/04/16 00:05:39 kernel.go:55: installed kernel "linux-6.19.11"
```

Build a disk image from rootfs and chroot

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl disk -N test-vm -C openrc -R alpine -s 8192
+ rmdir /var/run/firecracker/openrc.chroot
+ mkdir -p /var/run/firecracker/openrc.chroot
+ umount /var/run/firecracker/openrc.chroot
+ true
+ dd 'if=/dev/zero' 'of=/var/lib/firecracker/disk/test-vm.img' 'bs=1M' 'count=8192'
8192+0 records in
8192+0 records out
8589934592 bytes (8.0GB) copied, 20.115676 seconds, 407.2MB/s
+ mkfs.ext4 /var/lib/firecracker/disk/test-vm.img
mke2fs 1.47.3 (8-Jul-2025)
Discarding device blocks: done
Creating filesystem with 2097152 4k blocks and 524288 inodes
Filesystem UUID: 674a3ba2-3cc6-42ce-858a-2835f0a93e6e
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632

Allocating group tables: done
Writing inode tables: done
Creating journal (16384 blocks): done
Writing superblocks and filesystem accounting information: done

+ mount /var/lib/firecracker/disk/test-vm.img /var/run/firecracker/openrc.chroot
+ tar -C /var/run/firecracker/openrc.chroot -xzvf /var/lib/firecracker/conf/alpine.rootfs.tar.gz
+ cp -av /var/lib/firecracker/conf/openrc.chroot.sh /var/run/firecracker/openrc.chroot/chroot.sh
'/var/lib/firecracker/conf/openrc.chroot.sh' -> '/var/run/firecracker/openrc.chroot/chroot.sh'
+ chroot /var/run/firecracker/openrc.chroot sh -ex chroot.sh
+ echo 'nameserver 1.1.1.1'
+ ln -s /var/cache/apk /etc/apk/cache
+ apk update
v3.23.3-535-g80547603fef [https://dl-cdn.alpinelinux.org/alpine/v3.23/main]
v3.23.4-2-g66056411722 [https://dl-cdn.alpinelinux.org/alpine/v3.23/community]
OK: 27578 distinct packages available
+ apk add openrc alpine-base git tar grep curl less socat openssh openssl iproute2 nftables
OK: 34.8 MiB in 72 packages
+ apk cache sync -v
Downloading 16 packages...
+ mkdir -p /lib/modules
+ passwd -d root
passwd: password for root changed by root
+ echo firecracker
+ rc-update add hostname
 * service hostname added to runlevel sysinit
+ sed -i -E 's/^#ttyS0:/ttyS0:/' /etc/inittab
+ mkdir -m700 /root/.ssh
+ cat
+ rc-update add networking
 * service networking added to runlevel sysinit
+ rc-update add sshd
 * service sshd added to runlevel sysinit
+ rc-update add devfs boot
 * service devfs added to runlevel boot
+ rc-update add procfs boot
 * service procfs added to runlevel boot
+ rc-update add sysfs boot
 * service sysfs added to runlevel boot
2026/04/16 00:06:52 disk.go:51: installed disk "test-vm"
```

Create a shared directory (if you want to use virtiofsd)

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl share -N test-vm
2026/04/16 00:07:58 share.go:48: installed share /var/lib/firecracker/share/test-vm.share
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl share
test-vm	/var/lib/firecracker/share/test-vm.share
```

Create a tap interface for network

```
alpine-hypervisor:~/firecracker-tools# ip tuntap add tap0 mode tap
alpine-hypervisor:~/firecracker-tools# ip l set tap0 up
alpine-hypervisor:~/firecracker-tools# ip a add 10.128.128.129/30 dev tap0
alpine-hypervisor:~/firecracker-tools# ip a
10: tap0: <NO-CARRIER,BROADCAST,MULTICAST,UP> mtu 1500 qdisc pfifo_fast state DOWN group default qlen 1000
    link/ether 46:45:d2:72:3c:dc brd ff:ff:ff:ff:ff:ff
    inet 10.128.128.129/30 scope global tap0
       valid_lft forever preferred_lft forever
```

Create the vm using the disk, kernel, and share

```
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl vm -N test-vm -D test-vm -K linux-6.19.11 -m 8192 -mmds -S test-vm
2026/04/16 00:11:17 vm.go:165: WARNING! virtiofsd support is not yet a mainline firecracker feature. You will need to build Firecracker yourself from a fork that supports virtiofsd.
2026/04/16 00:11:17 vm.go:177: wrote config to /var/lib/firecracker/conf/test-vm.json
```

Boot the vm

```
alpine-hypervisor:~/firecracker-tools# source env.sh virtiofsd
alpine-hypervisor:~/firecracker-tools# ./fcctl/fcctl vm test-vm start
2026/04/16 00:13:01 vm.go:121: starting vm "test-vm" from "/var/lib/firecracker/conf/test-vm.json"
2026/04/16 00:13:01 vm.go:124: starting firecracker:
	/root/firecracker-tools/firecracker/firecracker-with-virtiofsd/build/cargo_target/debug/firecracker \
		--api-sock \
		/var/run/firecracker/test-vm.sock \
		--config-file \
		/var/lib/firecracker/conf/test-vm.json \
		--id \
		test-vm
2026/04/16 00:13:01 share.go:52: starting virtiofsd for "test-vm"...
+ /usr/libexec/virtiofsd '--socket-path=/var/run/firecracker/test-vm.test-vm.share.sock' '--tag=test-vm' '--shared-dir=/var/lib/firecracker/share/test-vm.share'
2026-04-16T00:13:02.005621633 [test-vm:main] Running Firecracker v1.16.0-dev
2026-04-16T00:13:02.007436718 [test-vm:main] Listening on API socket ("/var/run/firecracker/test-vm.sock").
2026-04-16T00:13:02.008161189 [test-vm:fc_api] API server started.
[2026-04-16T00:13:02Z INFO  virtiofsd] Waiting for vhost-user socket connection...
[2026-04-16T00:13:02Z INFO  virtiofsd] Client connected, servicing requests
[    0.000000] Linux version 6.19.11 (ajaso@treadstone) (gcc (GCC) 14.2.1 20250405, GNU ld (GNU Binutils) 2.44) #1 SMP Tue Apr 14 17:22:21 PDT 2026
[    0.000000] Command line: rootfstype=ext4 rw console=ttyS0 pci=off root=/dev/vda rw virtio_mmio.device=4K@0xc0001000:5 virtio_mmio.device=4K@0xc0002000:6 virtio_mmio.device=4K@0xc0003000:7
```
