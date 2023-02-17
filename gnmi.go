package agent

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/openconfig/gnmi/proto/gnmi"
	"github.com/openconfig/gnmic/utils"
)

type SystemInfo struct {
	Name                string
	Version             string
	ChassisType         string
	ChassisMacAddress   string
	ChassisCLEICode     string
	ChassisPartNumber   string
	ChassisSerialNumber string
}

type MgmtAddresses struct {
	IPv4Address net.IP
	IPv4Prefix  uint8
	IPv6Address net.IP
	IPv6Prefix  uint8
}

var sysInfoPaths = []*gnmi.Path{
	{
		Elem: []*gnmi.PathElem{
			{Name: "system"},
			{Name: "name"},
			{Name: "host-name"},
		},
	},
	{
		Elem: []*gnmi.PathElem{
			{Name: "interface",
				Key: map[string]string{"name": "mgmt0"},
			},
			{Name: "subinterface"},
			{Name: "ipv4"},
			{Name: "address"},
			{Name: "status"},
		},
	},
	{
		Elem: []*gnmi.PathElem{
			{Name: "interface",
				Key: map[string]string{"name": "mgmt0"},
			},
			{Name: "subinterface"},
			{Name: "ipv6"},
			{Name: "address"},
			{Name: "status"},
		},
	},
	{
		Elem: []*gnmi.PathElem{
			{Name: "system"},
			{Name: "information"},
			{Name: "version"},
		},
	},
	{
		Elem: []*gnmi.PathElem{
			{Name: "platform"},
			{Name: "chassis"},
		},
	},
}

var mgmtIPPaths = []*gnmi.Path{
	{
		Elem: []*gnmi.PathElem{
			{Name: "interface",
				Key: map[string]string{"name": "mgmt0"},
			},
			{Name: "subinterface"},
			{Name: "ipv4"},
			{Name: "address"},
			{Name: "status"},
		},
	},
	{
		Elem: []*gnmi.PathElem{
			{Name: "interface",
				Key: map[string]string{"name": "mgmt0"},
			},
			{Name: "subinterface"},
			{Name: "ipv6"},
			{Name: "address"},
			{Name: "status"},
		},
	},
}

func (a *Agent) GetSystemInfo(ctx context.Context) (*SystemInfo, error) {
	sctx, cancel := context.WithCancel(ctx)
	defer cancel()
	rsp, err := a.Target.Get(sctx,
		&gnmi.GetRequest{
			Path:     sysInfoPaths,
			Type:     gnmi.GetRequest_STATE,
			Encoding: gnmi.Encoding_ASCII,
		})
	if err != nil {
		return nil, err
	}
	sysInfo := new(SystemInfo)
	for _, n := range rsp.GetNotification() {
		for _, u := range n.GetUpdate() {
			path := utils.GnmiPathToXPath(u.GetPath(), true)
			if strings.Contains(path, "system/name") {
				sysInfo.Name = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "system/information/version") {
				sysInfo.Version = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "platform/chassis/type") {
				sysInfo.ChassisType = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "platform/chassis/mac-address") {
				sysInfo.ChassisMacAddress = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "platform/chassis/part-number") {
				sysInfo.ChassisPartNumber = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "platform/chassis/clei-code") {
				sysInfo.ChassisCLEICode = u.GetVal().GetStringVal()
			}
			if strings.Contains(path, "platform/chassis/serial-number") {
				sysInfo.ChassisSerialNumber = u.GetVal().GetStringVal()
			}
		}
	}
	a.logger.Printf("system info: %+v", sysInfo)
	return sysInfo, nil
}

func (a *Agent) GetMgmtAddresses(ctx context.Context) (*MgmtAddresses, error) {
	sctx, cancel := context.WithCancel(ctx)
	defer cancel()

	rsp, err := a.Target.Get(sctx,
		&gnmi.GetRequest{
			Path:     mgmtIPPaths,
			Type:     gnmi.GetRequest_STATE,
			Encoding: gnmi.Encoding_ASCII,
		})
	if err != nil {
		return nil, err
	}
	mgmtIPs := new(MgmtAddresses)

	for _, n := range rsp.GetNotification() {
		for _, u := range n.GetUpdate() {
			path := utils.GnmiPathToXPath(u.GetPath(), true)
			if strings.HasPrefix(path, "interface") {
				if strings.Contains(path, "/ipv4/address/status") {
					ip := getPathKeyVal(u.GetPath(), "address", "ip-prefix")
					ipPrefix := strings.Split(ip, "/")
					if len(ipPrefix) != 2 {
						continue
					}
					mgmtIPs.IPv4Address = net.ParseIP(ipPrefix[0])
					prefix, _ := strconv.Atoi(ipPrefix[1])
					mgmtIPs.IPv4Prefix = uint8(prefix)
				}
				if strings.Contains(path, "/ipv6/address/status") {
					ip := getPathKeyVal(u.GetPath(), "address", "ip-prefix")
					ipPrefix := strings.Split(ip, "/")
					if len(ipPrefix) != 2 {
						continue
					}
					mgmtIPs.IPv6Address = net.ParseIP(ipPrefix[0])
					prefix, _ := strconv.Atoi(ipPrefix[1])
					mgmtIPs.IPv6Prefix = uint8(prefix)
				}
			}
		}
	}
	a.logger.Printf("mgmtIPs: %+v", mgmtIPs)
	return mgmtIPs, nil
}

func getPathKeyVal(p *gnmi.Path, elem, key string) string {
	for _, e := range p.GetElem() {
		if e.Name == elem {
			return e.Key[key]
		}
	}
	return ""
}
