#!/bin/sh -ex

# dns config
echo 'nameserver 1.1.1.1' > /etc/resolv.conf

# base packages
ln -s /var/cache/apk /etc/apk/cache
apk update
apk add \
    openrc \
    alpine-base \
    git \
    tar \
    grep \
    curl \
    less \
    socat \
    openssh \
    openssl \
    iproute2 \
    nftables
apk cache sync -v

# prepare kernel modules dir
mkdir -p /lib/modules

# user config
passwd -d root
echo 'firecracker' > /etc/hostname
rc-update add hostname
sed -i -E 's/^#ttyS0:/ttyS0:/' /etc/inittab
mkdir -m700 ~/.ssh || ls -l ~/.ssh
#curl -so ~/.ssh/authorized_keys https://github.com/youruser.keys
#chmod 600 ~/.ssh/authorized_keys

# networking config
cat >/etc/network/interfaces <<EOF
auto lo
iface lo inet loopback

auto eth0
iface eth0 inet static
        address 10.128.128.130
        netmask 255.255.255.252
        gateway 10.128.128.129
EOF
rc-update add networking
rc-update add sshd

rc-update add devfs boot
rc-update add procfs boot
rc-update add sysfs boot
