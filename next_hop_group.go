package agent

import (
	"context"
	"fmt"

	"github.com/nokia/srlinux-ndk-go/v21/ndk"
)

func (a *Agent) NxHopGrpAddOrUpdate(ctx context.Context, nhGroupInfo ...*ndk.NextHopGroupInfo) error {
	resp, err := a.NextHopGroupServiceClient.NextHopGroupAddOrUpdate(ctx, &ndk.NextHopGroupRequest{
		GroupInfo: nhGroupInfo,
	})
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) NxHopGrpDelete(ctx context.Context, nhGroupKey ...*ndk.NextHopGroupKey) error {
	resp, err := a.NextHopGroupServiceClient.NextHopGroupDelete(ctx, &ndk.NextHopGroupDeleteRequest{
		GroupKey: nhGroupKey,
	})
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) NxHopGrpSyncStart(ctx context.Context) error {
	resp, err := a.NextHopGroupServiceClient.SyncStart(ctx, new(ndk.SyncRequest))
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}

func (a *Agent) NxHopGrpSyncEnd(ctx context.Context) error {
	resp, err := a.NextHopGroupServiceClient.SyncEnd(ctx, new(ndk.SyncRequest))
	if err != nil {
		return err
	}
	if resp.GetStatus() == 0 {
		return nil
	}
	return fmt.Errorf("%s: %s", resp.GetStatus(), resp.GetErrorStr())
}
