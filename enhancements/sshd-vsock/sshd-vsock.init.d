#!/sbin/openrc-run

description="Forward VSOCK port 22 to sshd on localhost TCP port 22"

name="sshd-vsock"
command="socat"
command_args="VSOCK-LISTEN:22,fork TCP:localhost:22"
pidfile="/run/sshd-vsock.pid"
command_background=true

depend() {
        need sshd
        # verify that we have VSOCKET support
        gzip -dc /proc/config.gz | grep -qm1 VIRTIO_VSOCKETS=y
        ls -l /dev/vsock >/dev/null
}
