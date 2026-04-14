SHELL := /bin/bash

TARGETS = fcctl firecracker kernel rootfs

$(TARGETS):
	make -C $(@) build

ALPINE_VER ?= 3.23.3
KERNEL_VER ?= 6.19.11
PROFILE ?= openrc
TAP_DEV ?= tap0
TAP_ADDR ?= 10.128.128.129/30

FIRECRACKER = ./firecracker/firecracker/build/cargo_target/debug/firecracker
FCCTL = ./fcctl/fcctl
CHROOT = alpine-$(PROFILE)
KERNEL = linux-$(KERNEL_VER)
ROOTFS = alpine-minirootfs-$(ALPINE_VER)-x86_64
DISK = demo-alpine-$(ALPINE_VER)-$(PROFILE)
VM = demo-alpine-$(subst .,-,$(ALPINE_VER))-$(PROFILE)

CONF_DIR = /var/lib/firecracker
VM_F = $(CONF_DIR)/conf/$(VM).json
CHROOT_F = $(CONF_DIR)/conf/$(CHROOT).chroot.sh
ROOTFS_F = $(CONF_DIR)/conf/$(ROOTFS).rootfs.tar.gz
KERNEL_F = $(CONF_DIR)/kernel/$(KERNEL).vmlinux
DISK_F = $(CONF_DIR)/disk/$(DISK).img

$(FCCTL): fcctl

$(FIRECRACKER): firecracker

$(CHROOT_F): $(FCCTL) rootfs/chroot_openrc.sh
	$(FCCTL) chroot \
		-N alpine-openrc \
		-f rootfs/chroot_openrc.sh

$(KERNEL_F): $(FCCTL) kernel/linux-$(KERNEL_VER)-vmlinux kernel/linux-$(KERNEL_VER)-config
	$(FCCTL) kernel \
		-N $(KERNEL) \
		-kernel kernel/linux-$(KERNEL_VER)-vmlinux \
		-config kernel/linux-$(KERNEL_VER)-config

rootfs/$(ROOTFS).tar.gz:
	make -C rootfs download-minitrootfs

$(ROOTFS_F): rootfs/$(ROOTFS).tar.gz $(FCCTL)
	$(FCCTL) rootfs \
		-N $(ROOTFS) \
		-f $(<)

$(DISK_F): $(CHROOT_F) $(ROOTFS_F) $(FCCTL) 
	$(FCCTL) disk \
		-N $(DISK) \
		-c $(CHROOT) \
		-r $(ROOTFS) \
		-s 2048

$(VM_F): $(KERNEL_F) $(DISK_F) $(FCCTL)
	$(FCCTL) vm \
		-N $(VM) \
		-D $(DISK) \
		-K $(KERNEL) \
		-mmds \
		-tap-device $(TAP_DEV)/aa:bb:cc:dd:ee:ff

init-demo: $(VM_F)

start-demo: $(FCCTL) $(FIRECRACKER)
	ip tuntap add $(TAP_DEV) mode tap ||:
	ip link set $(TAP_DEV) up
	ip addr add $(TAP_ADDR) dev $(TAP_DEV) 2>/dev/null ||:
	export FC_BIN=$(FIRECRACKER) && \
		$(FCCTL) vm $(VM) start

stop-demo: $(FCCTL)
	$(FCCTL) vm $(VM) stop

.PHONY: $(TARGETS) init-demo start-demo
