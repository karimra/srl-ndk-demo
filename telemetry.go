package agent

import (
	"context"
	"fmt"

	"github.com/nokia/srlinux-ndk-go/v21/ndk"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *Agent) TelemetryAddOrUpdate(ctx context.Context, jsPath string, jsData string) {
	telReq := &ndk.TelemetryUpdateRequest{
		State: []*ndk.TelemetryInfo{
			{
				Key: &ndk.TelemetryKey{
					JsPath: jsPath,
				},
				Data: &ndk.TelemetryData{
					JsonContent: jsData,
				},
			},
		},
	}
	a.logger.Printf("update telemetry request: %+v", telReq)
	b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(telReq)
	if err != nil {
		a.logger.Printf("telemetry request Marshal failed: %+v", err)
	} else {
		a.logger.Printf("telemetry update:\n%s\n", string(b))
	}

	r1, err := a.TelemetryServiceClient.TelemetryAddOrUpdate(ctx, telReq)
	if err != nil {
		a.logger.Printf("could not update telemetry key=%s: err=%v", jsPath, err)
		return
	}
	a.logger.Printf("Telemetry add/update status: %s, error_string: %q", r1.GetStatus().String(), r1.GetErrorStr())
}

func (a *Agent) TelemetryDelete(ctx context.Context, jsPath string) error {
	telReq := &ndk.TelemetryDeleteRequest{
		Key: []*ndk.TelemetryKey{
			{
				JsPath: jsPath,
			},
		},
	}

	b, err := prototext.MarshalOptions{Multiline: true, Indent: "  "}.Marshal(telReq)
	if err != nil {
		a.logger.Printf("telemetry request Marshal failed: %+v", err)
	} else {
		a.logger.Printf("telemetry delete:\n%s\n", string(b))
	}

	r, err := a.TelemetryServiceClient.TelemetryDelete(ctx, telReq)
	if err != nil {
		return fmt.Errorf("could not delete telemetry for key : %s", jsPath)
	}
	if r.GetStatus() != ndk.SdkMgrStatus_kSdkMgrSuccess {
		return fmt.Errorf("telemetry delete failed: %s", r.GetErrorStr())
	}
	return nil
}
