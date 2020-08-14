package agent

import (
	"context"
	"io"
	"log"
	"net"
	"time"

	ndk "github.com/karimra/go-srl-ndk"
)

func (a *Agent) StartConfigNotificationStream(ctx context.Context) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for config notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("config notification registration status : %s streamID %d", notificationResponse.Status, notificationResponse.GetStreamId())

	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:       ndk.NotificationRegisterRequest_AddSubscription,
		StreamId: notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_Config{ // config
			Config: &ndk.ConfigSubscriptionRequest{},
		},
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartNwInstNotificationStream(ctx context.Context) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for NwInst notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("NwInst notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())

	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:       ndk.NotificationRegisterRequest_AddSubscription,
		StreamId: notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_NwInst{ // NwInst
			NwInst: &ndk.NetworkInstanceSubscriptionRequest{},
		},
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartInterfaceNotificationStream(ctx context.Context, ifName string) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for Intf notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("interface notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.Status == ndk.SdkMgrStatus_kSdkMgrFailed {
		log.Printf("interface notification subscribe failed")
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}

	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_Intf{},
	}
	if ifName != "" {
		key := &ndk.InterfaceKey{
			IfName: ifName,
		}
		notificationRegisterRequest.SubscriptionTypes = &ndk.NotificationRegisterRequest_Intf{
			Intf: &ndk.InterfaceSubscriptionRequest{
				Key: key,
			},
		}
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartLLDPNeighNotificationStream(ctx context.Context, ifName, chassisType, chassisID string) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for Intf notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("LLDPNeighbor notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.Status == ndk.SdkMgrStatus_kSdkMgrFailed {
		log.Printf("LLDPNeighbor notification subscribe failed")
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_LldpNeighbor{},
	}

	if ifName != "" || chassisID != "" || chassisType != "" {
		key := &ndk.LldpNeighborKeyPb{
			InterfaceName: ifName,
			// ChassisId:     chassisID,
			// ChassisType: ndk.LldpNeighborKeyPb_CHASSIS_COMPONENT,
		}
		notificationRegisterRequest = &ndk.NotificationRegisterRequest{
			Op:       ndk.NotificationRegisterRequest_AddSubscription,
			StreamId: notificationResponse.GetStreamId(),
			SubscriptionTypes: &ndk.NotificationRegisterRequest_LldpNeighbor{ // LLDPNeigh
				LldpNeighbor: &ndk.LldpNeighborSubscriptionRequest{
					Key: key,
				},
			},
		}
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartBFDSessionNotificationStream(ctx context.Context, srcIP, dstIP net.IP, instance *uint32) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for Intf notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("BFDSession notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.Status == ndk.SdkMgrStatus_kSdkMgrFailed {
		log.Printf("BFDSession notification subscribe failed")
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	bfdSession := &ndk.BfdSessionSubscriptionRequest{
		Key: &ndk.BfdmgrGeneralSessionKeyPb{},
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_BfdSession{},
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
	if srcIP != nil || dstIP != nil || instance != nil {
		notificationRegisterRequest = &ndk.NotificationRegisterRequest{
			Op:       ndk.NotificationRegisterRequest_AddSubscription,
			StreamId: notificationResponse.GetStreamId(),
			SubscriptionTypes: &ndk.NotificationRegisterRequest_BfdSession{ // BFDSession
				BfdSession: bfdSession,
			},
		}
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartRouteNotificationStream(ctx context.Context, netInstance string, ipAddr net.IP, prefixLen uint32) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for Intf notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("Route notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.Status == ndk.SdkMgrStatus_kSdkMgrFailed {
		log.Printf("Route notification subscribe failed")
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}

	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:       ndk.NotificationRegisterRequest_AddSubscription,
		StreamId: notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_Route{ // route
			Route: &ndk.IpRouteSubscriptionRequest{
				//Key: key,
			},
		},
	}
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
		notificationRegisterRequest = &ndk.NotificationRegisterRequest{
			Op:       ndk.NotificationRegisterRequest_AddSubscription,
			StreamId: notificationResponse.GetStreamId(),
			SubscriptionTypes: &ndk.NotificationRegisterRequest_Route{ // route
				Route: &ndk.IpRouteSubscriptionRequest{
					Key: key,
				},
			},
		}
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) StartAppIdNotificationStream(ctx context.Context, id *uint32) chan *ndk.NotificationStreamResponse {
CREATESUB:
	// get subscription and streamID
	notificationResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx,
		&ndk.NotificationRegisterRequest{
			Op: ndk.NotificationRegisterRequest_Create,
		})
	if err != nil {
		log.Printf("agent %s could not register for Intf notifications: %v", a.Name, err)
		log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	log.Printf("AppId notification registration status: %s, subscriptionID=%d, streamID=%d",
		notificationResponse.Status, notificationResponse.GetSubId(), notificationResponse.GetStreamId())
	if notificationResponse.Status == ndk.SdkMgrStatus_kSdkMgrFailed {
		log.Printf("AppId notification subscribe failed")
		time.Sleep(a.RetryTimer)
		goto CREATESUB
	}
	notificationRegisterRequest := &ndk.NotificationRegisterRequest{
		Op:                ndk.NotificationRegisterRequest_AddSubscription,
		StreamId:          notificationResponse.GetStreamId(),
		SubscriptionTypes: &ndk.NotificationRegisterRequest_Appid{},
	}
	if id != nil {
		notificationRegisterRequest = &ndk.NotificationRegisterRequest{
			Op:       ndk.NotificationRegisterRequest_AddSubscription,
			StreamId: notificationResponse.GetStreamId(),
			SubscriptionTypes: &ndk.NotificationRegisterRequest_Appid{ // AppId
				Appid: &ndk.AppIdentSubscriptionRequest{
					Key: &ndk.AppIdentKey{Id: *id},
				},
			},
		}
	}
	return a.startNotificationStream(ctx, notificationRegisterRequest, notificationResponse.GetSubId())
}
func (a *Agent) startNotificationStream(ctx context.Context, req *ndk.NotificationRegisterRequest, subID uint64) chan *ndk.NotificationStreamResponse {
	streamChan := make(chan *ndk.NotificationStreamResponse)
	log.Printf("starting stream with req=%+v", req)
	go func() {
		defer close(streamChan)
		defer func() {
			log.Printf("agent %s deleting subscription %d", a.Name, subID)
			a.SdkMgrService.Client.NotificationRegister(context.TODO(), &ndk.NotificationRegisterRequest{
				Op:    ndk.NotificationRegisterRequest_DeleteSubscription,
				SubId: subID,
			})
		}()
	GETSTREAM:
		registerResponse, err := a.SdkMgrService.Client.NotificationRegister(ctx, req)
		if err != nil {
			log.Printf("agent %s failed registering to notification with req=%+v: %v", a.Name, req, err)
			log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
			time.Sleep(a.RetryTimer)
			goto GETSTREAM
		}
		if registerResponse.GetStatus() == ndk.SdkMgrStatus_kSdkMgrFailed {
			log.Printf("failed to get stream with req: %v", req)
			log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
			time.Sleep(a.RetryTimer)
			goto GETSTREAM
		}
		stream, err := a.NotificationService.Client.NotificationStream(ctx,
			&ndk.NotificationStreamRequest{
				StreamId: req.GetStreamId(),
			})
		if err != nil {
			log.Printf("agent %s failed creating stream client with req=%+v: %v", a.Name, req, err)
			log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
			time.Sleep(a.RetryTimer)
			goto GETSTREAM
		}

		for {
			// select {
			// case <-ctx.Done():
			// 	return
			// default:
			ev, err := stream.Recv()
			if err == io.EOF {
				log.Printf("agent %s received EOF for stream %v", a.Name, req.GetSubscriptionTypes())
				log.Printf("agent %s retrying in %s", a.Name, a.RetryTimer)
				time.Sleep(a.RetryTimer)
				goto GETSTREAM
			}
			if err != nil {
				log.Printf("agent %s failed to receive notification: %v", a.Name, err)
				continue
			}
			streamChan <- ev
		}
		//}
	}()
	return streamChan
}
