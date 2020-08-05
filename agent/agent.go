package agent

import (
	ndk "github.com/karimra/go-srl-ndk"
	"google.golang.org/grpc"
)

type Agent struct {
	AppID uint32

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
}
