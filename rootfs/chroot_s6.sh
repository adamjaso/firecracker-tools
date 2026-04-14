#!/bin/sh -ex

# dns config
echo 'nameserver 1.1.1.1' > /etc/resolv.conf

# base packages
ln -s /var/cache/apk /etc/apk/cache
apk update
apk add \
    alpine-base \
    s6-overlay \
    agetty \
    less \
    curl \
    ifupdown-ng \
    openssh \
    iproute2 \
    nftables
apk cache sync -v

# s6 mods
mkdir -p /etc/services.d
mkdir -p /etc/s6-linux-init/current/run-image
cp -av /usr/bin/s6-* /etc/s6-linux-init/current/run-image

# s6 mounts service (supports ssh pty)
mkdir -p /etc/cont-init.d
cat >/etc/cont-init.d/00-setup <<EOF
#!/bin/sh
#mount -t devtmpfs devtmpfs /dev
mount -t proc proc /proc
mkdir -p /dev/pts
mount -t devpts devpts /dev/pts
rm /dev/ptmx
ln -s /dev/pts/ptmx /dev/ptmx
hostname -F /etc/hostname
ifup -a
EOF
chmod +x /etc/cont-init.d/00-setup


# s6 getty (login prompt)
mkdir -p /etc/services.d/getty
cat >/etc/services.d/getty/run <<EOF
#!/bin/sh
exec /sbin/agetty -n -L 115200 ttyS0 vt100
#exec /sbin/getty -n -l /bin/sh 115200 ttyS0
EOF
chmod +x /etc/services.d/getty/run

# s6 sshd
mkdir -p /etc/services.d/sshd/log
cat >/etc/services.d/sshd/run <<EOF
#!/bin/sh
ssh-keygen -A 
exec /usr/sbin/sshd -D
EOF
chmod +x /etc/services.d/sshd/run

# s6 sshd logging
cat >/etc/services.d/sshd/log/run <<EOF
#!/bin/sh
mkdir -p /var/log/sshd
exec s6-log -b T n20 s1000000 /var/log/sshd
EOF
chmod +x /etc/services.d/sshd/log/run

# user config
passwd -d root
echo ttyS0 > /etc/securetty
echo 'firecracker' > /etc/hostname
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

mkdir -p /etc/services.d
cat > /etc/s6-linux-init/current/run-image/init <<EOF
#!/bin/sh
#exec >/dev/console 2>&1
exec s6-svscan /etc/services.d
EOF
chmod +x /etc/s6-linux-init/current/run-image/init
