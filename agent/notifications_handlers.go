package agent

import ndk "github.com/karimra/go-srl-ndk"

func (a *Agent) HandleNotification(notif *ndk.Notification) error {
	if notif == nil {
		return nil
	}
	switch notif := notif.SubscriptionTypes.(type) {
	case *ndk.Notification_Appid:
		return a.handleAppId(notif.Appid)
	case *ndk.Notification_Config:
		return a.handleConfig(notif.Config)
	case *ndk.Notification_Intf:
		return a.handleIntf(notif.Intf)
	case *ndk.Notification_NwInst:
		return a.handleNwInst(notif.NwInst)
	case *ndk.Notification_LldpNeighbor:
		return a.handleNwLldpNeighbor(notif.LldpNeighbor)
	case *ndk.Notification_BfdSession:
		return a.handleBfdSession(notif.BfdSession)
	case *ndk.Notification_Route:
		return a.handleRoute(notif.Route)
	}
	return nil
}

func (a *Agent) handleAppId(n *ndk.AppIdentNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.m.Lock()
		defer a.m.Unlock()
		a.AppIDs[n.GetKey().Id] = n.GetData()
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleConfig(n *ndk.ConfigNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleIntf(n *ndk.InterfaceNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.m.Lock()
		defer a.m.Unlock()
		a.Intf[n.GetKey().IfName] = n.GetData()
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleNwInst(n *ndk.NetworkInstanceNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
		a.m.Lock()
		defer a.m.Unlock()
		a.NwInst[n.GetKey().InstName] = n.GetData()
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleNwLldpNeighbor(n *ndk.LldpNeighborNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleBfdSession(n *ndk.BfdSessionNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
func (a *Agent) handleRoute(n *ndk.IpRouteNotification) error {
	switch n.GetOp() {
	case ndk.SdkMgrOperation_Create:
	case ndk.SdkMgrOperation_Delete:
	case ndk.SdkMgrOperation_Change:
	}
	return nil
}
