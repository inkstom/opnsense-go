package ipsec

import (
	"context"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestIPsecSettings(t *testing.T) {
	c := newPoolTestController()
	ctx := context.Background()

	before, err := c.SettingsGet(ctx)
	if err != nil {
		t.Fatalf("SettingsGet: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c.SettingsSet(ctx, &before.Ipsec)
		_ = c.Api.ReconfigureService(ctx, "/ipsec/service/reconfigure")
	})

	want := before.Ipsec
	want.General.DisableVpnRules = "1"
	want.General.PassthroughNetworks = api.SelectedMapList{"192.0.2.0/24", "198.51.100.0/24"}
	if res, err := c.SettingsSet(ctx, &want); err != nil || res.Result != "saved" {
		t.Fatalf("SettingsSet = %+v, %v", res, err)
	}
	if err := c.Api.ReconfigureService(ctx, "/ipsec/service/reconfigure"); err != nil {
		t.Fatalf("reconfigure: %v", err)
	}
	got, err := c.SettingsGet(ctx)
	g := got.Ipsec.General
	if err != nil || g.DisableVpnRules != "1" || len(g.PassthroughNetworks) != 2 || len(g.UserSource) != 0 {
		t.Fatalf("SettingsGet after set = %+v, %v", g, err)
	}
}
