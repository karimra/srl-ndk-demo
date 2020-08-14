package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	ndk "github.com/karimra/go-srl-ndk"
	"google.golang.org/grpc"
)

var grpcAddress = "localhost:50053"

var retryTimeout = 5 * time.Second

type HandleFunc func(context.Context, *ndk.NotificationStreamResponse)

type Agent struct {
	Name       string
	RetryTimer time.Duration
	AppID      uint32

	GRPCConn *grpc.ClientConn

	SdkMgrService struct {
		Client ndk.SdkMgrServiceClient
	}
	NotificationService struct {
		Client ndk.SdkNotificationServiceClient
	}
	TelemetryService struct {
		Client ndk.SdkMgrTelemetryServiceClient
	}
	RouteService struct {
		Client ndk.SdkMgrRouteServiceClient
	}
	MPLSRouteService struct {
		Client ndk.SdkMgrMplsRouteServiceClient
	}
	NextHopGroupService struct {
		Client ndk.SdkMgrNextHopGroupServiceClient
	}

	m            *sync.RWMutex
	Config       map[string]*ndk.ConfigData           // keys is jsPath
	NwInst       map[string]*ndk.NetworkInstanceData  // keys is instance name
	Intf         map[string]*ndk.InterfaceData        // keys is interface name
	AppIDs       map[uint32]*ndk.AppIdentData         // keys is appIdent
	LLDPNeighbor map[string][]*ndk.LldpNeighborDataPb // key is interfaceName, if there is a neighbor, there is a local interface
	BFDSession   *BfdSession
	IPRoute      map[string]map[string]*ndk.RoutePb // instanceName / ipAddr/prefixLen
}

func NewAgent(ctx context.Context, name string) (*Agent, error) {
	a := &Agent{
		m:            new(sync.RWMutex),
		Config:       make(map[string]*ndk.ConfigData),
		NwInst:       make(map[string]*ndk.NetworkInstanceData),
		Intf:         make(map[string]*ndk.InterfaceData),
		AppIDs:       make(map[uint32]*ndk.AppIdentData),
		LLDPNeighbor: make(map[string][]*ndk.LldpNeighborDataPb),
		BFDSession:   newBfdSession(),
		IPRoute:      make(map[string]map[string]*ndk.RoutePb),
	}
	a.Name = name
	a.RetryTimer = retryTimeout

	var err error
	a.GRPCConn, err = grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("grpc dial failed: %v", err)
		return nil, err
	}
	a.SdkMgrService.Client = ndk.NewSdkMgrServiceClient(a.GRPCConn)

	nctx, cancel := context.WithTimeout(ctx, a.RetryTimer)
	defer cancel()
	r, err := a.SdkMgrService.Client.AgentRegister(nctx, &ndk.AgentRegistrationRequest{})
	if err != nil {
		return nil, fmt.Errorf("agent %s registration failed: %v", a.Name, err)
	}
	a.AppID = r.GetAppId()
	log.Printf("agent %s: registration status: %v", a.Name, r.GetStatus())
	log.Printf("agent %s: registration appID: %v", a.Name, r.GetAppId())
	// create telemetry and notifications Clients
	a.TelemetryService.Client = ndk.NewSdkMgrTelemetryServiceClient(a.GRPCConn)
	a.NotificationService.Client = ndk.NewSdkNotificationServiceClient(a.GRPCConn)
	return a, nil
}

func (a *Agent) KeepAlive(ctx context.Context, period time.Duration) {
	newTicker := time.NewTicker(period)
	for {
		select {
		case <-newTicker.C:
			keepAliveResponse, err := a.SdkMgrService.Client.KeepAlive(ctx, &ndk.KeepAliveRequest{})
			if err != nil {
				log.Printf("agent %s: failed to send keep alive request: %v", a.Name, err)
				continue
			}
			log.Printf("agent %s: received keepAliveResponse, status=%v", a.Name, keepAliveResponse.Status)
		case <-ctx.Done():
			log.Printf("agent %s: received %v, shutting down keepAlives", a.Name, ctx.Err())
			newTicker.Stop()
			return
		}
	}
}
