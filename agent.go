package agent

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/nokia/srlinux-ndk-go/ndk"
	"github.com/openconfig/gnmic/target"
	"github.com/openconfig/gnmic/types"
	"google.golang.org/grpc"
)

var grpcAddress = "localhost:50053"
var gnmiUnixAddress = "unix:///opt/srlinux/var/run/sr_gnmi_server"
var defaultRetryTimeout = 5 * time.Second

type NotificationService struct {
	Client   ndk.SdkNotificationServiceClient
	Handlers *NotificationHandlers
}

type Agent struct {
	Name         string
	retryTimeout time.Duration
	AppID        uint32
	// gRPC connection used to connect to the NDK server
	gRPCConn *grpc.ClientConn
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
	TelemetryServiceClient ndk.SdkMgrTelemetryServiceClient
	RouteServiceClient     ndk.SdkMgrRouteServiceClient
	// MPLSRouteServiceClient    ndk.SdkMgrMplsRouteServiceClient // deprecated
	NextHopGroupServiceClient ndk.SdkMgrNextHopGroupServiceClient

	m            *sync.RWMutex
	Config       map[string]*ndk.ConfigData          // key is jsPath
	NwInst       map[string]*ndk.NetworkInstanceData // key is instance name
	Intf         map[string]*ndk.InterfaceData       // key is interface name
	AppIDs       map[uint32]*ndk.AppIdentData        // key is appIdent
	LLDPNeighbor map[string]*ndk.LldpNeighborDataPb  // key is interfaceName, if there is a neighbor, there is a local interface
	BFDSession   *BfdSession                         // TODO: ??
	IPRoute      map[string]map[string]*ndk.RoutePb  // instanceName / ipAddr/prefixLen
	NhGroup      map[uint64]*ndk.NextHopGroup        // instanceName / Name
	// config transactions cache
	configTrx []*ndk.ConfigNotification
	// logger
	logger *log.Logger
}

func New(ctx context.Context, name string, opts ...agentOption) (*Agent, error) {
	a := &Agent{
		retryTimeout: defaultRetryTimeout,
		logger:       log.New(os.Stderr, "", log.LstdFlags),
		//
		m:            new(sync.RWMutex),
		Config:       make(map[string]*ndk.ConfigData),
		NwInst:       make(map[string]*ndk.NetworkInstanceData),
		Intf:         make(map[string]*ndk.InterfaceData),
		AppIDs:       make(map[uint32]*ndk.AppIdentData),
		LLDPNeighbor: make(map[string]*ndk.LldpNeighborDataPb),
		BFDSession:   newBfdSession(),
		IPRoute:      make(map[string]map[string]*ndk.RoutePb),
		NhGroup:      make(map[uint64]*ndk.NextHopGroup),
		configTrx:    make([]*ndk.ConfigNotification, 0),
	}
	a.Name = name
	for _, opt := range opts {
		opt(a)
	}
	var err error
	a.gRPCConn, err = grpc.Dial(grpcAddress, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc dial failed: %v", err)
	}
	a.SdkMgrServiceClient = ndk.NewSdkMgrServiceClient(a.gRPCConn)

	nctx, cancel := context.WithTimeout(ctx, a.retryTimeout)
	defer cancel()
	r, err := a.SdkMgrServiceClient.AgentRegister(nctx, &ndk.AgentRegistrationRequest{})
	if err != nil {
		return nil, fmt.Errorf("agent %q registration failed: %v", a.Name, err)
	}
	a.AppID = r.GetAppId()
	a.logger.Printf("agent %s: registration status: %v", a.Name, r.GetStatus())
	a.logger.Printf("agent %s: registration appID: %v", a.Name, r.GetAppId())
	// create telemetry and notifications Clients
	a.TelemetryServiceClient = ndk.NewSdkMgrTelemetryServiceClient(a.gRPCConn)
	a.NotificationService.Client = ndk.NewSdkNotificationServiceClient(a.gRPCConn)
	return a, nil
}

type agentOption func(*Agent)

func WithLogger(l *log.Logger) agentOption {
	return func(a *Agent) {
		a.logger = l
		if a.logger == nil {
			a.logger = log.New(os.Stderr, "", log.LstdFlags)
		}
	}
}

func WithRetryTimer(r time.Duration) agentOption {
	return func(a *Agent) {
		a.retryTimeout = r
		if a.retryTimeout <= 0 {
			a.retryTimeout = defaultRetryTimeout
		}
	}
}

func (a *Agent) CreateGNMIClient(ctx context.Context, tc *types.TargetConfig) error {
	if tc == nil {
		tc = new(types.TargetConfig)
	}
	if tc.Address == "" {
		tc.Address = gnmiUnixAddress
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
				a.logger.Printf("agent %q: failed to send keep alive request: %v", a.Name, err)
				continue
			}
			a.logger.Printf("agent %q: received keepAliveResponse, status=%v", a.Name, keepAliveResponse.Status)
		case <-ctx.Done():
			a.logger.Printf("agent %q: received %v, shutting down keepAlives", a.Name, ctx.Err())
			newTicker.Stop()
			return
		}
	}
}
