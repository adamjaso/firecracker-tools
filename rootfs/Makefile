SHELL := /bin/bash

PROFILE ?= openrc
SIZE ?= 4096
ALPINE_VER ?= 3.23.3
CHROOT_SCRIPT ?= chroot_$(PROFILE).sh

PREFIX = alpine-$(PROFILE)-$(ALPINE_VER)-build
DISK = $(PREFIX).img
ALPINE_MMVER = $(basename $(ALPINE_VER))
ROOTFS_DIR = alpine-minirootfs-$(ALPINE_VER)-x86_64
ROOTFS_TAR = $(ROOTFS_DIR).tar.gz
ROOTFS_URL = https://dl-cdn.alpinelinux.org/alpine/v$(ALPINE_MMVER)/releases/x86_64/$(ROOTFS)

$(ROOTFS_TAR):
	wget -O$(@) $(ROOTFS_URL)

%-build: $(ROOTFS_TAR)
	mkdir -p $(@)
	tar -C $(@) -xzvf $(<)

%-build/chroot.sh: $(CHROOT_SCRIPT) %-build
	cp -a $(<) $(@)

%.tar.gz: %/chroot.sh
	chroot $(*) sh -ex chroot.sh
	tar --same-owner --same-permissions -C $(*) -czf $(@) .

%.img: %.tar.gz
	dd if=/dev/zero of=$(@) bs=1M count=$(SIZE)
	mkfs.ext4 $(@)
	mkdir -p $(*)
	mount $(@) $(*)
	tar --same-owner --same-permissions -C $(*) -xzf $(<)
	umount $(*)

download-minirootfs: $(ROOTFS_TAR)
	# no-op

build: $(DISK)
	# no-op

clean:
	rm -r $(DISK) $(PREFIX)-build

.PHONY: clean build download-minirootfs
