package util

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var space = regexp.MustCompile(`\s+`)

type (
	Brs map[string]*Br
	Br  struct {
		ID   string
		Name string
		Ifs  []string
	}
	Addrs []Addr
	Addr  struct {
		Family string `json:"family"`
		Addr   string `json:"local"`
		Prefix int    `json:"prefixlen"`
	}
	Ifs []If
	If  struct {
		Name   string   `json:"ifname"`
		Flags  []string `json:"flags"`
		Master string   `json:"master"` // contains bridge interface name for tap interfaces

		// operstate, address, link_type, linkinfo.info_kind are only valid for "ip link"
		State string `json:"operstate"`
		Mac   string `json:"address"`
		Type  string `json:"link_type"`
		Info  struct {
			Kind string `json:"info_kind"`
		} `json:"linkinfo,omitempty"`

		// processes is only valid for "ip tuntap" output
		Procs []struct {
			Name string `json:"name"`
			Pid  int    `json:"pid"`
		} `json:"processes"`

		// addr_info is only valid for "ip addr"
		Addrs Addrs `json:"addr_info"`

		// an index to simplify checking for presence of a flag
		flagset map[string]struct{} `json:"-"`
	}
)

func (a Addr) String() string {
	return fmt.Sprintf("%s/%d", a.Addr, a.Prefix)
}

func (a Addrs) First() Addr {
	if len(a) > 0 {
		return a[0]
	}
	return Addr{}
}

func (a Addrs) ByAddr(ifaddr string) Addr {
	for _, a := range a {
		if a.Addr == ifaddr || a.String() == ifaddr {
			return a
		}
	}
	return Addr{}
}

func (i If) HasFlag(flag string) bool {
	_, ok := i.flagset[flag]
	return ok
}

func (i If) InetAddrs() Addrs {
	addrs := Addrs{}
	for _, a := range i.Addrs {
		if a.Family == "inet" {
			addrs = append(addrs, a)
		}
	}
	return addrs
}

func (i If) GetAddr() (If, error) {
	ifs, err := GetAddrs(i.Name)
	return ifs.First(), err
}

func (i If) GetLink() (If, error) {
	ifs, err := GetLinks(i.Name)
	return ifs.First(), err
}

func (i If) GetMaster() (If, error) {
	m, err := GetAddrs(i.Master)
	return m.ByName(i.Master), err
}

func GetTuntaps(name string) (Ifs, error) {
	return listIproute2("tuntap", name)
}

func GetLinks(name string) (Ifs, error) {
	return listIproute2("link", name)
}

func GetAddrs(name string) (Ifs, error) {
	return listIproute2("addr", name)
}

func listIproute2(kind, name string) (Ifs, error) {
	args := []string{"-j", "-d", kind}
	if name != "" {
		args = append(args, "show", name)
	}
	res := Ifs{}
	resbytes := &bytes.Buffer{}
	log.Println("ip", args)
	cmd := exec.Command("ip", args...)
	cmd.Stdout = resbytes
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if cmd.ProcessState.ExitCode() == 1 {
			// exit 1 means not found
			return res, nil
		}
		return res, err
	} else if err := json.NewDecoder(resbytes).Decode(&res); err != nil {
		return res, err
	}
	return res.applyFlagsets(), nil
}

func (ifs Ifs) First() If {
	if len(ifs) > 0 {
		return ifs[0]
	}
	return If{}
}

func (ifs Ifs) applyFlagsets() Ifs {
	for i := range ifs {
		ifs[i].flagset = map[string]struct{}{}
		for _, f := range ifs[i].Flags {
			ifs[i].flagset[f] = struct{}{}
		}
	}
	return ifs
}

func (ifs Ifs) ByName(name string) If {
	for _, i := range ifs {
		if i.Name == name {
			return i
		}
	}
	return If{}
}

func (ifs Ifs) ByState(state string) Ifs {
	res := Ifs{}
	for _, i := range ifs {
		if i.State == state {
			res = append(res, i)
		}
	}
	return res
}

func (ifs Ifs) ByFlag(flag string) Ifs {
	res := Ifs{}
	for _, i := range ifs {
		if i.HasFlag(flag) {
			res = append(res, i)
		}
	}
	return res
}

func (ifs Ifs) Bridges() Ifs {
	res := Ifs{}
	for _, i := range ifs {
		if i.Type == "bridge" {
			res = append(res, i)
		}
	}
	return res
}

func (ifs Ifs) ByType(kind string) Ifs {
	res := Ifs{}
	for _, i := range ifs {
		if kind == "tap" && i.HasFlag("tap") {
			res = append(res, i)
		} else if i.Type == kind || i.Info.Kind == kind {
			res = append(res, i)
		}
	}
	return res
}

func addIf(ifname, iftype string) error {
	var cmd *exec.Cmd
	if iftype == "tap" || iftype == "tun" {
		cmd = exec.Command("ip", "tuntap", "add", ifname, "mode", iftype)
	} else {
		cmd = exec.Command("ip", "link", "add", ifname, "type", iftype)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddIf(ifname, iftype string) (If, error) {
	log.Printf("addif %s type %s", ifname, iftype)
	if ifs, err := GetLinks(ifname); err != nil {
		return If{}, err
	} else if if_ := ifs.ByName(ifname); if_.Name != "" {
		return if_, nil
	} else if err := addIf(ifname, iftype); err != nil {
		return If{}, err
	} else if ifs, err := GetLinks(ifname); err != nil {
		return If{}, err
	} else {
		return ifs.ByName(ifname), nil
	}
}

func SetIfUp(ifname string) error {
	log.Printf("setifup %s", ifname)
	cmd := exec.Command("ip", "link", "set", ifname, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddAddr(ifname, ifaddr string) error {
	addrs, err := GetAddrs(ifname)
	if err != nil {
		return err
	}
	addr := addrs.ByName(ifname).Addrs.ByAddr(ifaddr)
	if addr.Addr != "" {
		return nil
	}
	cmd := exec.Command("ip", "addr", "add", ifaddr, "dev", ifname)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func addBr(brname, ifname string) error {
	cmd := exec.Command("brctl", "addif", brname, ifname)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func AddBr(brname, ifname string) error {
	log.Printf("addbr %s %s", brname, ifname)
	if brs, err := GetBridges(brname); err != nil {
		return err
	} else if br, ok := brs[brname]; ok {
		for _, i := range br.Ifs {
			if i == ifname {
				return nil
			}
		}
	}
	return addBr(brname, ifname)
}

func GetBridges(name string) (Brs, error) {
	args := []string{"show"}
	if name != "" {
		args = append(args, name)
	}
	resbytes := &bytes.Buffer{}
	cmd := exec.Command("brctl", args...)
	cmd.Stdout = resbytes
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	brname := ""
	res := Brs{}
	lines := bufio.NewScanner(resbytes)
	for lines.Scan() {
		if strings.HasPrefix(lines.Text(), "bridge name") {
			// skip header
			continue
		} else if strings.HasPrefix(lines.Text(), " ") {
			ifname := strings.TrimSpace(lines.Text())
			res[brname].Ifs = append(res[brname].Ifs, ifname)

		} else if parts := space.Split(lines.Text(), 4); len(parts) == 4 {
			brname = parts[0]
			res[brname] = &Br{ID: parts[1], Name: brname, Ifs: []string{strings.TrimSpace(parts[3])}}

		} else if len(parts) == 3 {
			res[brname] = &Br{ID: parts[1], Name: brname, Ifs: []string{}}
		}
	}
	return res, nil
}
