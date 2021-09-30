package agent

import (
	"context"

	ndk "github.com/karimra/go-srl-ndk"
)

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
		return a.handleNwLldpNeighbor(ctx, notif.LldpNeighbor)
	case *ndk.Notification_BfdSession:
		return a.handleBfdSession(ctx, notif.BfdSession)
	case *ndk.Notification_Route:
		return a.handleRoute(ctx, notif.Route)
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
	case ndk.SdkMgrOperation_Change:
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
func (a *Agent) handleIntf(ctx context.Context, n *ndk.InterfaceNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.Intf[n.GetKey().GetIfName()] = n.GetData()
		a.NotificationService.Handlers.InterfaceCreate(ctx, n)
	case ndk.SdkMgrOperation_Change:
		a.Intf[n.GetKey().GetIfName()] = n.GetData()
		a.NotificationService.Handlers.InterfaceChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.Intf, n.GetKey().GetIfName())
		a.NotificationService.Handlers.InterfaceDelete(ctx, n)
	}
	return nil
}

// Network Instance
func (a *Agent) handleNwInst(ctx context.Context, n *ndk.NetworkInstanceNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.NwInst[n.GetKey().GetInstName()] = n.GetData()
		a.NotificationService.Handlers.NetworkInstanceCreate(ctx, n)
	case ndk.SdkMgrOperation_Change:
		a.NwInst[n.GetKey().GetInstName()] = n.GetData()
		a.NotificationService.Handlers.NetworkInstanceChange(ctx, n)
	case ndk.SdkMgrOperation_Delete:
		delete(a.NwInst, n.GetKey().GetInstName())
		a.NotificationService.Handlers.NetworkInstanceDelete(ctx, n)
	}
	return nil
}

// LLDP Neighbors
func (a *Agent) handleNwLldpNeighbor(ctx context.Context, n *ndk.LldpNeighborNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}

// BFDSession
func (a *Agent) handleBfdSession(ctx context.Context, n *ndk.BfdSessionNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}

// Route
func (a *Agent) handleRoute(ctx context.Context, n *ndk.IpRouteNotification) error {
	a.m.Lock()
	defer a.m.Unlock()
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
