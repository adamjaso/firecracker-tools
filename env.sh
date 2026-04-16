# firecracker binary build
case "${1}" in
    virtiofsd) export FC_BIN=$PWD/firecracker/firecracker-with-virtiofsd/build/cargo_target/debug/firecracker ;;
    *)         export FC_BIN=$PWD/firecracker/firecracker/build/cargo_target/debug/firecracker ;;
esac
