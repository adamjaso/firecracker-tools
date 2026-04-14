# Forwarding SSH over Firecracker Vsock

Firecracker supports a [Vsock mechanism][^1] that enables host-guest
connectivity via a UNIX socket on the host side and a VSOCK socket on the guest side.

This Vsock is specified in the VM configuration file `--config-file`, under the
["vsock" attribute][^2].

This feature can be used by the host for guest management i.e. think [qemu-guest-agent][^3].

## Rationale

As I was browsing for ways to share a host directory with the guest, I encountered an article that mentioned
"ssh over VSOCK". [^4] Which made me curious to see what could be done with the Firecracker Vsock feature.

After some digging, I was able to get ssh over VSOCK working with my Firecracker VM.

## How to

Connecting to the guest from the host via the firecracker `uds_path` UNIX socket requires a
preliminary `CONNECT <GUEST_TCP_PORT>\n` text command to initiate the connection to VSOCK CID
(CID is a port number, essentially the VSOCK analog for TCP/UDP port). I used `guest_cid` of `22` so
the text command was `CONNECT 22\n`.

After initiating the `CONNECT`, then you can use the connection as a tunnel to any guest service that is
listening on `CID 22`. To make this work, you can run `socat` as a listener on the VSOCK `VSOCK-LISTEN:22,fork`, and forward all
traffic to `sshd` listening on `TCP:localhost:22`. The [sshd-vsock.init.d](./sshd-vsock.init.d) script is an OpenRC init script
that does this.

After starting the `sshd-vsock` command `socat VSOCK-LISTEN:22,fork TCP:localhost:22` on the guest. You're
ready to ssh from the host to the guest.

Using `ProxyCommand sh -c '(echo CONNECT 22; cat) | nc -U /var/run/firecracker/demo-alpine-3.23.3-openrc.vsock22.sock'`
with `ssh`, you initiate the connection to CID 22 with `echo`, and then read from stdin with `cat` and pipe it all to
`nc` connected to the host side VSOCK UNIX socket.

You may need to run `ssh` as root if the VSOCK is owned by root and your user is not authorized.

You will also need to configure `sshd` appropriately to allow you to log in as the user of your choice, i.e. `.ssh/authorized_keys`.

```
user@host ~/firecracker-tools/enhancements/sshd-vsock $ doas ssh -oIdentityAgent=$SSH_AUTH_SOCK -F ssh_config root@firecracker-vsock
Warning: Permanently added 'firecracker-vsock' (ED25519) to the list of known hosts.
Welcome to Alpine!

The Alpine Wiki contains a large amount of how-to guides and general
information about administrating Alpine systems.
See <https://wiki.alpinelinux.org/>.

You can setup the system with the command: setup-alpine

You may change this message by editing /etc/motd.

firecracker:~#
Connection to firecracker-vsock closed.
```

[^1]: https://github.com/firecracker-microvm/firecracker/blob/main/docs/vsock.md
[^2]: https://pkg.go.dev/github.com/firecracker-microvm/firecracker-go-sdk/client/models#Vsock
[^3]: https://www.qemu.org/docs/master/interop/qemu-ga.html
[^4]: https://virtio-win.github.io/Knowledge-Base/SSH-over-VSock.html
