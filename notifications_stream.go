package agent

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/nokia/srlinux-ndk-go/v21/ndk"
)

func (a *Agent) createNotificationSubscription(ctx context.Context) (uint64, uint64) {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrServiceClient.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		a.logger.Printf("agent %q could not register for notifications: %v", a.Name, err)
		a.logger.Printf("agent %q retrying in %s", a.Name, a.retryTimeout)
		time.Sleep(a.retryTimeout)
		goto CREATESUB
	}
	if notificationResponse.GetStatus() != ndk.SdkMgrStatus_kSdkMgrSuccess {
		a.logger.Printf("notification subscribe failed")
		time.Sleep(a.retryTimeout)
		goto CREATESUB
	}
	return notificationResponse.GetSubId(), notificationResponse.GetStreamId()
}

func (a *Agent) startNotificationStream(ctx context.Context, req *ndk.NotificationRegisterRequest, subID uint64, streamChan chan *ndk.NotificationStreamResponse) {
	a.logger.Printf("starting stream with req=%+v", req)
	defer close(streamChan)
	defer func() {
		a.logger.Printf("agent %s deleting subscription %d", a.Name, subID)
		a.SdkMgrServiceClient.NotificationRegister(context.TODO(), &ndk.NotificationRegisterRequest{
			Op:    ndk.NotificationRegisterRequest_DeleteSubscription,
			SubId: subID,
		})
	}()
GETSTREAM:
	registerResponse, err := a.SdkMgrServiceClient.NotificationRegister(ctx, req)
	if err != nil {
		a.logger.Printf("agent %s failed registering to notification with req=%+v: %v", a.Name, req, err)
		a.logger.Printf("agent %s retrying in %s", a.Name, a.retryTimeout)
		time.Sleep(a.retryTimeout)
		goto GETSTREAM
	}
	if registerResponse.GetStatus() == ndk.SdkMgrStatus_kSdkMgrFailed {
		a.logger.Printf("failed to get stream with req: %v", req)
		a.logger.Printf("agent %s retrying in %s", a.Name, a.retryTimeout)
		time.Sleep(a.retryTimeout)
		goto GETSTREAM
	}
	stream, err := a.NotificationService.Client.NotificationStream(ctx,
		&ndk.NotificationStreamRequest{
			StreamId: req.GetStreamId(),
		})
	if err != nil {
		a.logger.Printf("agent %s failed creating stream client with req=%+v: %v", a.Name, req, err)
		a.logger.Printf("agent %s retrying in %s", a.Name, a.retryTimeout)
		time.Sleep(a.retryTimeout)
		goto GETSTREAM
	}
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ev, err := stream.Recv()
			if err == io.EOF {
				a.logger.Printf("agent %s received EOF for stream %v", a.Name, req.GetSubscriptionTypes())
				a.logger.Printf("agent %s retrying in %s", a.Name, a.retryTimeout)
				time.Sleep(a.retryTimeout)
				goto GETSTREAM
			}
			if err != nil {
				a.logger.Printf("agent %s failed to receive notification: %v", a.Name, err)
				continue
			}
			streamChan <- ev
		}
	}
}

func (a *Agent) StartConfigNotificationStream(ctx context.Context) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("Config notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:       ndk.NotificationRegisterRequest_AddSubscription,
		StreamId: streamID,
		SubscriptionTypes: &ndk.NotificationRegisterRequest_Config{ // config
			Config: new(ndk.ConfigSubscriptionRequest),
		},
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartNwInstNotificationStream(ctx context.Context) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("NwInst notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:       ndk.NotificationRegisterRequest_AddSubscription,
		StreamId: streamID,
		SubscriptionTypes: &ndk.NotificationRegisterRequest_NwInst{ // NwInst
			NwInst: new(ndk.NetworkInstanceSubscriptionRequest),
		},
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartInterfaceNotificationStream(ctx context.Context, ifName string) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("Intf notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	subType := new(ndk.NotificationRegisterRequest_Intf)
	if ifName != "" {
		subType.Intf = &ndk.InterfaceSubscriptionRequest{
			Key: &ndk.InterfaceKey{
				IfName: ifName,
			},
		}
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          streamID,
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartLLDPNeighNotificationStream(ctx context.Context, ifName, chassisType, chassisID string) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("LLDPNeighbor notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	subType := new(ndk.NotificationRegisterRequest_LldpNeighbor)
	if ifName != "" || chassisID != "" || chassisType != "" {
		subType.LldpNeighbor = &ndk.LldpNeighborSubscriptionRequest{
			Key: &ndk.LldpNeighborKeyPb{
				InterfaceName: ifName,
				ChassisId:     chassisID,
			},
		}
		if v, ok := ndk.LldpNeighborKeyPb_ChassisIdType_value[chassisType]; ok {
			subType.LldpNeighbor.Key.ChassisType = ndk.LldpNeighborKeyPb_ChassisIdType(v)
		}
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          streamID,
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartBFDSessionNotificationStream(ctx context.Context, srcIP, dstIP net.IP, instance *uint32) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("BFDSession notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	bfdSession := &ndk.BfdSessionSubscriptionRequest{
		Key: new(ndk.BfdmgrGeneralSessionKeyPb),
	}
	if srcIP != nil {
		bfdSession.Key.SrcIpAddr = &ndk.IpAddressPb{Addr: srcIP}
	}
	if dstIP != nil {
		bfdSession.Key.DstIpAddr = &ndk.IpAddressPb{Addr: dstIP}
	}
	if instance != nil {
		bfdSession.Key.InstanceId = *instance
	}
	subType := new(ndk.NotificationRegisterRequest_BfdSession)
	if srcIP != nil || dstIP != nil || instance != nil {
		subType.BfdSession = bfdSession
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          streamID,
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartRouteNotificationStream(ctx context.Context, netInstance string, ipAddr net.IP, prefixLen uint32) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("Route notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	subType := new(ndk.NotificationRegisterRequest_Route)
	key := new(ndk.RouteKeyPb)
	if netInstance != "" {
		key.NetInstName = netInstance
	}
	if ipAddr != nil {
		key.IpPrefix = &ndk.IpAddrPrefLenPb{
			IpAddr:       &ndk.IpAddressPb{Addr: ipAddr},
			PrefixLength: prefixLen,
		}
	}
	if netInstance != "" || ipAddr != nil {
		subType.Route = &ndk.IpRouteSubscriptionRequest{
			Key: key,
		}
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          streamID,
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartAppIdNotificationStream(ctx context.Context, id *uint32) chan *ndk.NotificationStreamResponse {
	subID, streamID := a.createNotificationSubscription(ctx)
	a.logger.Printf("AppId notification registration: subscriptionID=%d, streamID=%d", subID, streamID)
	subType := new(ndk.NotificationRegisterRequest_Appid)
	if id != nil {
		subType.Appid = &ndk.AppIdentSubscriptionRequest{
			Key: &ndk.AppIdentKey{Id: *id},
		}
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          streamID,
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, subID, streamChan)
	return streamChan
}

func (a *Agent) StartNhGroupNotificationStream(ctx context.Context, instanceName, name string) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrServiceClient.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		a.logger.Printf("agent %s could not register for NHGroup notifications: %v", a.Name, err)
		a.logger.Printf("agent %s retrying in %s", a.Name, a.retryTimeout)
		time.Sleep(a.retryTimeout)
		goto CREATESUB
	}
	a.logger.Printf("NHGroup notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.GetStatus(), notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.GetStatus() != ndk.SdkMgrStatus_kSdkMgrSuccess {
		a.logger.Printf("NHGroup notification subscribe failed")
		time.Sleep(a.retryTimeout)
		goto CREATESUB
	}

	subType := new(ndk.NotificationRegisterRequest_Nhg)
	if name != "" && instanceName != "" {
		subType.Nhg = &ndk.NextHopGroupSubscriptionRequest{
			Key: &ndk.NextHopGroupKey{
				Name:                name,
				NetworkInstanceName: instanceName,
			},
		}
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          notificationResponse.GetStreamId(),
		SubscriptionTypes: subType,
	}
	streamChan := make(chan *ndk.NotificationStreamResponse)
	go a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId(), streamChan)
	return streamChan
}
