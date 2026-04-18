package fclib

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"fcctl/util"

	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

type (
	NetMember struct {
		VmName   string `json:"vm_name"`
		GuestMac string `json:"guest_mac"`
		HostTap  string `json:"host_tap"`
		Address  string `json:"address"`

		sortIndex int64
	}
	Net struct {
		Name           string   `json:"name"`
		Subnet         string   `json:"subnet"`
		GuestMacPrefix string   `json:"guest_mac_prefix"`
		Vms            []string `json:"vms"`

		Brname  string `json:"-"`
		File    string `json:"-"`
		Address string `json:"-"`
		members []NetMember
		subnet  *net.IPNet
		id      int64
	}
	Networks struct {
		Nets []*Net
	}
)

func ScanNets(dir string) (*Networks, error) {
	if dir == "" {
		dir = DefaultNetdir
	}
	networks := &Networks{Nets: []*Net{}}
	files, _ := filepath.Glob(fmt.Sprintf("%s/*.json", dir))
	for _, fn := range files {
		net := &Net{File: fn}
		if err := net.Read(); err != nil {
			return nil, err
		}
		networks.Nets = append(networks.Nets, net)
	}
	return networks, nil
}

func NewNet(name string) *Net {
	return &Net{
		Name: name,
		File: fmt.Sprintf("%s/%s.json", DefaultNetdir, name),
		Vms:  []string{},
	}
}

func (n *Net) CheckFile() error {
	return checkFile(n.File, "net file", ErrFileExists, nil)
}

func (n *Net) GetMembers() []NetMember {
	return n.members
}

func (n *Net) SetupBridge() error {
	if _, err := util.AddIf(n.Brname, "bridge"); err != nil {
		return err
	} else if err := util.SetIfUp(n.Brname); err != nil {
		return err
	} else if err := util.AddAddr(n.Brname, n.Address); err != nil {
		return err
	}
	for _, mem := range n.members {
		if _, err := util.AddIf(mem.HostTap, "tap"); err != nil {
			return err
		} else if err := util.SetIfUp(mem.HostTap); err != nil {
			return err
		}
	}
	for _, mem := range n.members {
		if err := util.AddBr(n.Brname, mem.HostTap); err != nil {
			return err
		}
	}
	return nil
}

func (n *Net) Write() error {
	f, err := os.OpenFile(n.File, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(n)
}

func (n *Net) Read() error {
	f, err := os.Open(n.File)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := json.NewDecoder(f).Decode(n); err != nil {
		return err
	}
	n.Brname = "fc-" + n.Name
	_, n.subnet, err = net.ParseCIDR(n.Subnet)
	if err != nil {
		return err
	}
	n.Address = n.AddrFromSubnet(1)
	n.members = make([]NetMember, len(n.Vms))
	for i, vmName := range n.Vms {
		n.members[i] = NetMember{
			VmName:   vmName,
			Address:  n.AddrFromSubnet(i + 2),
			GuestMac: fmt.Sprintf("%s:%02x", n.GuestMacPrefix, i),
			HostTap:  fmt.Sprintf("fctap%dm%d", n.id, i),
		}
	}
	parts := strings.Split(n.GuestMacPrefix, ":")
	if len(parts) != 5 {
		return fmt.Errorf("invalid guest mac prefix: %s", n.GuestMacPrefix)
	}
	n.id, err = strconv.ParseInt(parts[4], 16, 64)
	if err != nil {
		return err
	}
	return nil
}

func (n *Net) AddrFromSubnet(add int) string {
	ipint := uint32(0)
	addr := n.subnet.IP.To4()
	for i := 0; i < 4; i++ {
		ipint += uint32(addr[i]) << ((3 - i) * 8)
	}
	if add == -1 {
		add, _ = n.subnet.Mask.Size()
		add -= 1
	}
	ipint += uint32(add)
	newip := net.IPv4(0, 0, 0, 0).To4()
	for i := 3; i >= 0; i-- {
		newip[i] = byte(ipint % 256)
		ipint = ipint >> 8
	}
	return newip.String()
}

func (n *Networks) GetVmNetworkInterface(vmName string) []*models.NetworkInterface {
	ifaces := []*models.NetworkInterface{}
	for _, bnet := range n.Nets {
		for _, mem := range bnet.members {
			if mem.VmName == vmName {
				ifaces = append(ifaces, &models.NetworkInterface{
					IfaceID:     new(bnet.Brname),
					HostDevName: new(mem.HostTap),
					GuestMac:    mem.GuestMac,
				})
			}
		}
	}
	return ifaces
}

func GetBridgeNetworks(allVms []VmConf) ([]*Net, error) {
	bridgeNetworks := map[string]*Net{}
	vmConf := map[string]*util.VmConfFile{}
	vmTaps := map[string]util.Ifs{}
	vmBridges := map[string]util.Ifs{}
	for _, vm := range allVms {
		conf, err := vm.ReadVm()
		if err != nil {
			return nil, err
		}
		vmConf[vm.Name] = conf
	}

	// all bridges
	allAddrs, err := util.GetAddrs("")
	if err != nil {
		return nil, err
	}

	// map taps to vm confs
	for _, vm := range allVms {
		vmConf := vmConf[vm.Name]
		for _, dev := range vmConf.NetworkInterfaces {
			if _, ok := vmTaps[vm.Name]; !ok {
				vmTaps[vm.Name] = util.Ifs{}
			}
			vmTap := allAddrs.ByName(*dev.HostDevName)
			vmTaps[vm.Name] = append(vmTaps[vm.Name], vmTap)

			if _, ok := vmBridges[vm.Name]; !ok {
				vmBridges[vm.Name] = util.Ifs{}
			}
			vmBridge := allAddrs.ByName(vmTap.Master)
			vmBridges[vm.Name] = append(vmBridges[vm.Name], vmBridge)

			bridge, ok := bridgeNetworks[vmBridge.Name]
			if !ok {
				bridge = &Net{
					Brname:  vmBridge.Name,
					members: []NetMember{},
				}
				bridgeNetworks[vmBridge.Name] = bridge
			}
			bridge.members = append(bridge.members, NetMember{
				VmName:    vm.Name,
				HostTap:   vmTap.Name,
				GuestMac:  dev.GuestMac,
				sortIndex: vmConf.CreatedAt,
			})
		}
	}

	// sort bridges
	nets := []*Net{}
	for _, net := range bridgeNetworks {
		sort.SliceStable(net.members, func(i, j int) bool {
			return net.members[i].sortIndex < net.members[j].sortIndex
		})
		nets = append(nets, net)
	}
	sort.SliceStable(nets, func(i, j int) bool {
		return nets[i].Brname < nets[j].Brname
	})
	return nets, nil
}
