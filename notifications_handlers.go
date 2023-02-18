package agent

import (
	"context"
	"fmt"
	"sync"

	"github.com/nokia/srlinux-ndk-go/ndk"
)

type HandleFunc func(context.Context, *ndk.Notification) error
type AppIdHandleFunc func(context.Context, *ndk.AppIdentNotification) error
type ConfigHandleFunc func(context.Context, []*ndk.ConfigNotification) error
type InterfaceHandleFunc func(context.Context, *ndk.InterfaceNotification) error
type NwInstHandleFunc func(context.Context, *ndk.NetworkInstanceNotification) error
type LLDPNeighborHandleFunc func(context.Context, *ndk.LldpNeighborNotification) error
type BFDSessionHandleFunc func(context.Context, *ndk.BfdSessionNotification) error
type RouteHandleFunc func(context.Context, *ndk.IpRouteNotification) error
type NextHopGroupHandleFunc func(context.Context, *ndk.NextHopGroupNotification) error

type NotificationHandlers struct {
	// Default Handlers
	Create HandleFunc
	Change HandleFunc
	Delete HandleFunc
	// AppId Handlers
	AppIdCreate AppIdHandleFunc
	AppIdChange AppIdHandleFunc
	AppIdDelete AppIdHandleFunc
	// Config Handler
	ConfigHandler ConfigHandleFunc
	// Interface Handlers
	InterfaceCreate InterfaceHandleFunc
	InterfaceChange InterfaceHandleFunc
	InterfaceDelete InterfaceHandleFunc
	// Network Instance Handlers
	NetworkInstanceCreate NwInstHandleFunc
	NetworkInstanceChange NwInstHandleFunc
	NetworkInstanceDelete NwInstHandleFunc
	// LLDP Neighbors Handlers
	LLDPNeighborsCreate LLDPNeighborHandleFunc
	LLDPNeighborsChange LLDPNeighborHandleFunc
	LLDPNeighborsDelete LLDPNeighborHandleFunc
	// BFD Session Handlers
	BFDSessionCreate BFDSessionHandleFunc
	BFDSessionChange BFDSessionHandleFunc
	BFDSessionDelete BFDSessionHandleFunc
	// Route Handlers
	RouteCreate RouteHandleFunc
	RouteChange RouteHandleFunc
	RouteDelete RouteHandleFunc
	// NextHopGroup Handlers
	NextHopGroupCreate NextHopGroupHandleFunc
	NextHopGroupChange NextHopGroupHandleFunc
	NextHopGroupDelete NextHopGroupHandleFunc
}

func merge(ctx context.Context, cs ...<-chan *ndk.NotificationStreamResponse) <-chan *ndk.NotificationStreamResponse {
	out := make(chan *ndk.NotificationStreamResponse)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan *ndk.NotificationStreamResponse) {
			defer wg.Done()
			for v := range c {
				select {
				case out <- v:
				case <-ctx.Done():
					return
				}
			}
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (a *Agent) Start(ctx context.Context, streams ...<-chan *ndk.NotificationStreamResponse) {
	for {
		select {
		case ns := <-merge(ctx, streams...):
			for _, n := range ns.GetNotification() {
				a.HandleNotification(ctx, n)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (a *Agent) HandleNotification(ctx context.Context, notif *ndk.Notification) error {
	if notif == nil {
		return nil
	}
	switch notif := notif.SubscriptionTypes.(type) {
	case *ndk.Notification_Appid:
		a.handleAppId(ctx, notif.Appid)
	case *ndk.Notification_Config:
		return a.handleConfig(ctx, notif.Config)
	case *ndk.Notification_Intf:
		return a.handleIntf(ctx, notif.Intf)
	case *ndk.Notification_NwInst:
		return a.handleNwInst(ctx, notif.NwInst)
	case *ndk.Notification_LldpNeighbor:
		return a.handleLldpNeighbor(ctx, notif.LldpNeighbor)
	case *ndk.Notification_BfdSession:
		return a.handleBfdSession(ctx, notif.BfdSession)
	case *ndk.Notification_Route:
		return a.handleRoute(ctx, notif.Route)
	case *ndk.Notification_Nhg:
		return a.handleNextHopGroup(ctx, notif.Nhg)
	}
	return nil
}

// AppID
func (a *Agent) RegisterAppIdCreateNotificationHandler(h AppIdHandleFunc) {
	a.NotificationService.Handlers.AppIdCreate = h
}

func (a *Agent) RegisterAppIdChangeNotificationHandler(h AppIdHandleFunc) {
	a.NotificationService.Handlers.AppIdChange = h
}

func (a *Agent) RegisterAppIdDeleteNotificationHandler(h AppIdHandleFunc) {
	a.NotificationService.Handlers.AppIdDelete = h
}

func (a *Agent) handleAppId(ctx context.Context, n *ndk.AppIdentNotification) {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.AppIDs[n.GetKey().GetId()] = n.GetData()
		a.NotificationService.Handlers.AppIdCreate(ctx, n)
	case ndk.SdkMgrOperation_Update:
		a.AppIDs[n.GetKey().GetId()] = n.GetData()
		a.NotificationService.Handlers.AppIdChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.AppIDs, n.GetKey().GetId())
		a.NotificationService.Handlers.AppIdDelete(ctx, n)
	}
}

// Config
func (a *Agent) RegisterConfigNotificationsHandler(h ConfigHandleFunc) {
	a.NotificationService.Handlers.ConfigHandler = h
}

func (a *Agent) handleConfig(ctx context.Context, n *ndk.ConfigNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	if n.GetKey().GetJsPath() != ".commit.end" {
		a.configTrx = append(a.configTrx, n)
		return nil
	}
	defer func() { a.configTrx = make([]*ndk.ConfigNotification, 0) }()
	return a.NotificationService.Handlers.ConfigHandler(ctx, a.configTrx)
}

// Interface
func (a *Agent) RegisterInterfaceCreateNotificationHandler(h InterfaceHandleFunc) {
	a.NotificationService.Handlers.InterfaceCreate = h
}

func (a *Agent) RegisterInterfaceChangeNotificationHandler(h InterfaceHandleFunc) {
	a.NotificationService.Handlers.InterfaceChange = h
}

func (a *Agent) RegisterInterfaceDeleteNotificationHandler(h InterfaceHandleFunc) {
	a.NotificationService.Handlers.InterfaceDelete = h
}

func (a *Agent) handleIntf(ctx context.Context, n *ndk.InterfaceNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.Intf[n.GetKey().GetIfName()] = n.GetData()
		return a.NotificationService.Handlers.InterfaceCreate(ctx, n)
	case ndk.SdkMgrOperation_Update:
		a.Intf[n.GetKey().GetIfName()] = n.GetData()
		return a.NotificationService.Handlers.InterfaceChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.Intf, n.GetKey().GetIfName())
		return a.NotificationService.Handlers.InterfaceDelete(ctx, n)
	}
	return nil
}

// Network Instance
func (a *Agent) RegisterNwInstCreateNotificationHandler(h NwInstHandleFunc) {
	a.NotificationService.Handlers.NetworkInstanceCreate = h
}

func (a *Agent) RegisterNwInstChangeNotificationHandler(h NwInstHandleFunc) {
	a.NotificationService.Handlers.NetworkInstanceChange = h
}

func (a *Agent) RegisterNwInstDeleteNotificationHandler(h NwInstHandleFunc) {
	a.NotificationService.Handlers.NetworkInstanceDelete = h
}

func (a *Agent) handleNwInst(ctx context.Context, n *ndk.NetworkInstanceNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.NwInst[n.GetKey().GetInstName()] = n.GetData()
		return a.NotificationService.Handlers.NetworkInstanceCreate(ctx, n)
	case ndk.SdkMgrOperation_Update:
		a.NwInst[n.GetKey().GetInstName()] = n.GetData()
		return a.NotificationService.Handlers.NetworkInstanceChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.NwInst, n.GetKey().GetInstName())
		return a.NotificationService.Handlers.NetworkInstanceDelete(ctx, n)
	}
	return nil
}

// LLDP Neighbors
func (a *Agent) RegisterLldpNeighborCreateNotificationHandler(h LLDPNeighborHandleFunc) {
	a.NotificationService.Handlers.LLDPNeighborsCreate = h
}

func (a *Agent) RegisterLldpNeighborChangeNotificationHandler(h LLDPNeighborHandleFunc) {
	a.NotificationService.Handlers.LLDPNeighborsChange = h
}

func (a *Agent) RegisterLldpNeighborDeleteNotificationHandler(h LLDPNeighborHandleFunc) {
	a.NotificationService.Handlers.LLDPNeighborsDelete = h
}

func (a *Agent) handleLldpNeighbor(ctx context.Context, n *ndk.LldpNeighborNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		return a.NotificationService.Handlers.LLDPNeighborsCreate(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		return a.NotificationService.Handlers.LLDPNeighborsChange(ctx, n)
	case ndk.SdkMgrOperation_Update:
		return a.NotificationService.Handlers.LLDPNeighborsDelete(ctx, n)
	}
	return nil
}

// BFDSession
func (a *Agent) RegisterBFDSessionCreateNotificationHandler(h BFDSessionHandleFunc) {
	a.NotificationService.Handlers.BFDSessionCreate = h
}

func (a *Agent) RegisterBFDSessionChangeNotificationHandler(h BFDSessionHandleFunc) {
	a.NotificationService.Handlers.BFDSessionChange = h
}

func (a *Agent) RegisterBFDSessionDeleteNotificationHandler(h BFDSessionHandleFunc) {
	a.NotificationService.Handlers.BFDSessionDelete = h
}

func (a *Agent) handleBfdSession(ctx context.Context, n *ndk.BfdSessionNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		return a.NotificationService.Handlers.BFDSessionCreate(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		return a.NotificationService.Handlers.BFDSessionChange(ctx, n)
	case ndk.SdkMgrOperation_Update:
		return a.NotificationService.Handlers.BFDSessionDelete(ctx, n)
	}
	return nil
}

// Route
func (a *Agent) RegisterRouteCreateNotificationHandler(h RouteHandleFunc) {
	a.NotificationService.Handlers.RouteCreate = h
}

func (a *Agent) RegisterRouteChangeNotificationHandler(h RouteHandleFunc) {
	a.NotificationService.Handlers.RouteChange = h
}

func (a *Agent) RegisterRouteDeleteNotificationHandler(h RouteHandleFunc) {
	a.NotificationService.Handlers.RouteDelete = h
}

func (a *Agent) handleRoute(ctx context.Context, n *ndk.IpRouteNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		netIns := n.GetKey().GetNetInstName()
		ipAddr := n.GetKey().GetIpPrefix().GetIpAddr()
		ipPrefix := n.GetKey().GetIpPrefix().GetPrefixLength()
		a.IPRoute[netIns][fmt.Sprintf("%s/%d", ipAddr, ipPrefix)] = n.GetData()
		return a.NotificationService.Handlers.RouteCreate(ctx, n)
	case ndk.SdkMgrOperation_Update:
		netIns := n.GetKey().GetNetInstName()
		ipAddr := n.GetKey().GetIpPrefix().GetIpAddr()
		ipPrefix := n.GetKey().GetIpPrefix().GetPrefixLength()
		a.IPRoute[netIns][fmt.Sprintf("%s/%d", ipAddr, ipPrefix)] = n.GetData()
		return a.NotificationService.Handlers.RouteChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		netIns := n.GetKey().GetNetInstName()
		ipAddr := n.GetKey().GetIpPrefix().GetIpAddr()
		ipPrefix := n.GetKey().GetIpPrefix().GetPrefixLength()
		delete(a.IPRoute[netIns], fmt.Sprintf("%s/%d", ipAddr, ipPrefix))
		return a.NotificationService.Handlers.RouteDelete(ctx, n)
	}
	return nil
}

// NextHopGroup
func (a *Agent) RegisterNextHopGroupCreateNotificationHandler(h NextHopGroupHandleFunc) {
	a.NotificationService.Handlers.NextHopGroupCreate = h
}

func (a *Agent) RegisterNextHopGroupChangeNotificationHandler(h NextHopGroupHandleFunc) {
	a.NotificationService.Handlers.NextHopGroupChange = h
}

func (a *Agent) RegisterNextHopGroupDeleteNotificationHandler(h NextHopGroupHandleFunc) {
	a.NotificationService.Handlers.NextHopGroupDelete = h
}

func (a *Agent) handleNextHopGroup(ctx context.Context, n *ndk.NextHopGroupNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.NhGroup[n.GetKey()] = n.GetData()
		return a.NotificationService.Handlers.NextHopGroupCreate(ctx, n)
	case ndk.SdkMgrOperation_Update:
		a.NhGroup[n.GetKey()] = n.GetData()
		return a.NotificationService.Handlers.NextHopGroupChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.NhGroup, n.GetKey())
		return a.NotificationService.Handlers.NextHopGroupDelete(ctx, n)
	}
	return nil
}
