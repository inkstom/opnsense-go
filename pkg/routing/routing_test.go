package routing

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func newTestController() *Controller {
	return &Controller{Api: api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: true,
		MaxBackoff:    30,
		MinBackoff:    1,
		MaxRetries:    4,
	})}
}

func testGateway(name, ip string) *Gateway {
	return &Gateway{
		Disabled:       "0",
		Name:           name,
		Description:    "opnsense-go acceptance test",
		Interface:      "wan",
		IpProtocol:     "inet",
		Gateway:        ip,
		FarGateway:     "1", // the CI VM's WAN subnet is unknown; allow out-of-subnet gateways
		MonitorDisable: "1",
		Priority:       "200",
		Weight:         "1",
	}
}

func TestGateway(t *testing.T) {
	c := newTestController()
	ctx := context.Background()

	gw := testGateway("TF_TEST_GW1", "203.0.113.1")
	id, err := c.AddGateway(ctx, gw)
	if err != nil {
		t.Fatalf("AddGateway: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteGateway(ctx, id); err != nil {
			t.Errorf("DeleteGateway: %v", err)
		}
	})

	got, err := c.GetGateway(ctx, id)
	if err != nil {
		t.Fatalf("GetGateway: %v", err)
	}
	if got.Name != gw.Name || got.Gateway != gw.Gateway || string(got.Interface) != "wan" || string(got.IpProtocol) != "inet" {
		t.Fatalf("GetGateway mismatch: %+v", got)
	}

	gw.Priority = "150"
	gw.Description = "updated"
	if err := c.UpdateGateway(ctx, id, gw); err != nil {
		t.Fatalf("UpdateGateway: %v", err)
	}
	got, err = c.GetGateway(ctx, id)
	if err != nil {
		t.Fatalf("GetGateway after update: %v", err)
	}
	if got.Priority != "150" || got.Description != "updated" {
		t.Fatalf("update not applied: %+v", got)
	}
}

func TestGatewayGroup(t *testing.T) {
	c := newTestController()
	ctx := context.Background()

	gw1, err := c.AddGateway(ctx, testGateway("TF_TEST_GGW1", "203.0.113.11"))
	if err != nil {
		t.Fatalf("AddGateway 1: %v", err)
	}
	gw2, err := c.AddGateway(ctx, testGateway("TF_TEST_GGW2", "203.0.113.12"))
	if err != nil {
		_ = c.DeleteGateway(ctx, gw1)
		t.Fatalf("AddGateway 2: %v", err)
	}

	// Cleanups run last-in first-out: the group (registered below) is
	// deleted before its member gateways.
	t.Cleanup(func() {
		for _, g := range []string{gw1, gw2} {
			if err := c.DeleteGateway(ctx, g); err != nil {
				t.Errorf("DeleteGateway %s: %v", g, err)
			}
		}
	})

	group := &GatewayGroup{
		Name:        "TF_TEST_GROUP",
		Tier1:       api.SelectedMapList{"TF_TEST_GGW1"},
		Tier2:       api.SelectedMapList{"TF_TEST_GGW2"},
		Trigger:     "down",
		Description: "opnsense-go acceptance test",
	}

	// OPNsense caches the gateway option list (configd "interface gateways
	// list", cache_ttl 20s), so brand-new gateways are rejected with
	// "Option [...] not in list." for up to ~20 seconds.
	var id string
	deadline := time.Now().Add(60 * time.Second)
	for {
		id, err = c.AddGatewayGroup(ctx, group)
		if err == nil || !strings.Contains(err.Error(), "not in list") || time.Now().After(deadline) {
			break
		}
		time.Sleep(5 * time.Second)
	}
	if err != nil {
		t.Fatalf("AddGatewayGroup: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteGatewayGroup(ctx, id); err != nil {
			t.Errorf("DeleteGatewayGroup: %v", err)
		}
	})

	got, err := c.GetGatewayGroup(ctx, id)
	if err != nil {
		t.Fatalf("GetGatewayGroup: %v", err)
	}
	if got.Name != group.Name || got.Tier1.String() != "TF_TEST_GGW1" || got.Tier2.String() != "TF_TEST_GGW2" {
		t.Fatalf("GetGatewayGroup mismatch: %+v", got)
	}

	group.Tier1 = api.SelectedMapList{"TF_TEST_GGW1", "TF_TEST_GGW2"}
	group.Tier2 = api.SelectedMapList{}
	if err := c.UpdateGatewayGroup(ctx, id, group); err != nil {
		t.Fatalf("UpdateGatewayGroup: %v", err)
	}
	got, err = c.GetGatewayGroup(ctx, id)
	if err != nil {
		t.Fatalf("GetGatewayGroup after update: %v", err)
	}
	if len(got.Tier1) != 2 || len(got.Tier2) != 0 {
		t.Fatalf("update not applied: %+v", got)
	}
}
