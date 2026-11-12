package captiveportal

import (
	"context"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
	"github.com/browningluke/opnsense-go/pkg/interfaces"
)

// A zone needs a non-WAN interface; build a VLAN on the test VM's only NIC and
// assign it (the assignment API needs OPNsense 26.7.7 or later).
func TestZone(t *testing.T) {
	c := newTestController(t)
	ic := &interfaces.Controller{Api: c.Api}
	ctx := context.Background()

	vlanID, err := ic.AddVlan(ctx, &interfaces.Vlan{Parent: "vtnet0", Tag: "3904", Priority: "0", Device: "vlan0.3904", Description: "cp test"})
	if err != nil {
		t.Fatalf("AddVlan: %v", err)
	}
	t.Cleanup(func() { _ = ic.DeleteVlan(ctx, vlanID) })
	ifID, err := ic.AddAssignment(ctx, &interfaces.Assignment{Description: "CPTEST", Device: "vlan0.3904", Enable: "1", Lock: "0"})
	if err != nil {
		t.Skipf("interface assignment unavailable (OPNsense < 26.7.7?): %v", err)
	}
	t.Cleanup(func() { _ = ic.DeleteAssignment(ctx, ifID) })

	zoneID, err := c.AddZone(ctx, &Zone{Enabled: "0", Interfaces: api.SelectedMapList{ifID},
		AuthServers: api.SelectedMapList{"Local Database"}, AlwaysSendAccountingReqs: "0", IdleTimeout: "30",
		HardTimeout: "120", ConcurrentLogins: "1", AllowedAddresses: api.SelectedMapList{"192.0.2.10", "198.51.100.0/24"},
		AllowedMACAddresses: api.SelectedMapList{"00:11:22:33:44:55"}, ExtendedPreAuthData: "0", Description: "test zone"})
	if err != nil {
		t.Fatalf("AddZone: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteZone(ctx, zoneID) })

	z, err := c.GetZone(ctx, zoneID)
	if err != nil || len(z.AllowedAddresses) != 2 || z.Interfaces[0] != ifID || z.Template != "" || len(z.CorsOrigins) != 0 {
		t.Fatalf("GetZone = %+v, %v", z, err)
	}
	z.AllowedAddresses = api.SelectedMapList{}
	z.IdleTimeout = "60"
	if err := c.UpdateZone(ctx, zoneID, z); err != nil {
		t.Fatalf("UpdateZone: %v", err)
	}
	if z, err = c.GetZone(ctx, zoneID); err != nil || len(z.AllowedAddresses) != 0 || z.IdleTimeout != "60" {
		t.Fatalf("GetZone after update = %+v, %v", z, err)
	}
	if err := c.DeleteZone(ctx, zoneID); err != nil {
		t.Fatalf("DeleteZone: %v", err)
	}
}
