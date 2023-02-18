package agent

import (
	"context"
	"fmt"

	"github.com/nokia/srlinux-ndk-go/ndk"
)

func (a *Agent) RouteAddOrUpdate(ctx context.Context, routeInfo ...*ndk.RouteInfo) error {
	resp, err := a.RouteServiceClient.RouteAddOrUpdate(ctx, &ndk.RouteAddRequest{
		Routes: routeInfo,
	})
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) RouteDelete(ctx context.Context, routeKey ...*ndk.RouteKeyPb) error {
	in := &ndk.RouteDeleteRequest{
		Routes: routeKey,
	}
	resp, err := a.RouteServiceClient.RouteDelete(ctx, in)
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) RouteSyncStart(ctx context.Context) error {
	resp, err := a.RouteServiceClient.SyncStart(ctx, new(ndk.SyncRequest))
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) RouteSyncEnd(ctx context.Context) error {
	resp, err := a.RouteServiceClient.SyncEnd(ctx, new(ndk.SyncRequest))
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}
