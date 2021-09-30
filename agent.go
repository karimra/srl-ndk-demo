package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/karimra/gnmic/target"
	"github.com/karimra/gnmic/types"
	ndk "github.com/karimra/go-srl-ndk"
	"google.golang.org/grpc"
)

var grpcAddress = "localhost:50053"

var retryTimeout = 5 * time.Second

type HandleFunc func(context.Context, *ndk.Notification) error
type AppIdHandleFunc func(context.Context, *ndk.AppIdentNotification) error
type ConfigHandleFunc func(context.Context, []*ndk.ConfigNotification) error
type InterfaceHandleFunc func(context.Context, *ndk.InterfaceNotification) error
type NwInstHandleFunc func(context.Context, *ndk.NetworkInstanceNotification) error
type LLDPNeighborsHandleFunc func(context.Context, *ndk.LldpNeighborNotification) error
type BFDSessionHandleFunc func(context.Context, *ndk.BfdSessionNotification) error
type RouteHandleFunc func(context.Context, *ndk.IpRouteNotification) error

type Handlers struct {
	Create HandleFunc
	Change HandleFunc
	Delete HandleFunc
}

type NotificationHandlers struct {
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
	LLDPNeighborsCreate LLDPNeighborsHandleFunc
	LLDPNeighborsChange LLDPNeighborsHandleFunc
	LLDPNeighborsDelete LLDPNeighborsHandleFunc
	// BFD Session Handlers
	BFDSessionCreate BFDSessionHandleFunc
	BFDSessionChange BFDSessionHandleFunc
	BFDSessionDelete BFDSessionHandleFunc
	// Route Handlers
	RouteCreate RouteHandleFunc
	RouteChange RouteHandleFunc
	RouteDelete RouteHandleFunc
}

type NotificationService struct {
	Client   ndk.SdkNotificationServiceClient
	Handlers *NotificationHandlers
}

type Agent struct {
	Name       string
	RetryTimer time.Duration
	AppID      uint32
	// gRPC connection used to connect to the NDK server
	GRPCConn *grpc.ClientConn
	// unix socket gnmi client
	Target *target.Target
	// SDK Manager Client
	SdkMgrServiceClient ndk.SdkMgrServiceClient
	// Notifications Client and handlers
	NotificationService struct {
		Client   ndk.SdkNotificationServiceClient
		Handlers NotificationHandlers
	}
	//
	TelemetryServiceClient    ndk.SdkMgrTelemetryServiceClient
	RouteServiceClient        ndk.SdkMgrRouteServiceClient
	MPLSRouteServiceClient    ndk.SdkMgrMplsRouteServiceClient
	NextHopGroupServiceClient ndk.SdkMgrNextHopGroupServiceClient

	m            *sync.RWMutex
	Config       map[string]*ndk.ConfigData          // key is jsPath
	NwInst       map[string]*ndk.NetworkInstanceData // key is instance name
	Intf         map[string]*ndk.InterfaceData       // key is interface name
	AppIDs       map[uint32]*ndk.AppIdentData        // key is appIdent
	LLDPNeighbor map[string]*ndk.LldpNeighborDataPb  // key is interfaceName, if there is a neighbor, there is a local interface
	BFDSession   *BfdSession                         // TODO: ??
	IPRoute      map[string]map[string]*ndk.RoutePb  // instanceName / ipAddr/prefixLen
	// config transactions cache
	configTrx []*ndk.ConfigNotification
}

func NewAgent(ctx context.Context, name string, opts ...agentOption) (*Agent, error) {
	a := &Agent{
		m:            new(sync.RWMutex),
		Config:       make(map[string]*ndk.ConfigData),
		NwInst:       make(map[string]*ndk.NetworkInstanceData),
		Intf:         make(map[string]*ndk.InterfaceData),
		AppIDs:       make(map[uint32]*ndk.AppIdentData),
		LLDPNeighbor: make(map[string]*ndk.LldpNeighborDataPb),
		BFDSession:   newBfdSession(),
		IPRoute:      make(map[string]map[string]*ndk.RoutePb),
		configTrx:    make([]*ndk.ConfigNotification, 0),
	}
	for _, opt := range opts {
		opt(a)
	}
	a.Name = name
	a.RetryTimer = retryTimeout
	//
	var err error
	a.GRPCConn, err = grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		log.Printf("grpc dial failed: %v", err)
		return nil, err
	}
	a.SdkMgrServiceClient = ndk.NewSdkMgrServiceClient(a.GRPCConn)

	nctx, cancel := context.WithTimeout(ctx, a.RetryTimer)
	defer cancel()
	r, err := a.SdkMgrServiceClient.AgentRegister(nctx, &ndk.AgentRegistrationRequest{})
	if err != nil {
		return nil, fmt.Errorf("agent %s registration failed: %v", a.Name, err)
	}
	a.AppID = r.GetAppId()
	log.Printf("agent %s: registration status: %v", a.Name, r.GetStatus())
	log.Printf("agent %s: registration appID: %v", a.Name, r.GetAppId())
	// create telemetry and notifications Clients
	a.TelemetryServiceClient = ndk.NewSdkMgrTelemetryServiceClient(a.GRPCConn)
	a.NotificationService.Client = ndk.NewSdkNotificationServiceClient(a.GRPCConn)
	return a, nil
}

type agentOption func(*Agent)

func (a *Agent) CreateGNMIClient(ctx context.Context, tc *types.TargetConfig) error {
	if tc.Address == "" {
		tc.Address = "unix:///opt/srlinux/var/run/sr_gnmi_server"
	}
	if tc.Insecure == nil && tc.SkipVerify == nil {
		tc.Insecure = new(bool)
		*tc.Insecure = true
	}
	a.Target = target.NewTarget(tc)
	return a.Target.CreateGNMIClient(ctx)
}

func (a *Agent) KeepAlive(ctx context.Context, period time.Duration) {
	newTicker := time.NewTicker(period)
	for {
		select {
		case <-newTicker.C:
			keepAliveResponse, err := a.SdkMgrServiceClient.KeepAlive(ctx, &ndk.KeepAliveRequest{})
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
