package interfaces

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

// Requires OPNsense >= 26.7.7 (core interfaces/assignment API).
func TestAssignment(t *testing.T) {
	c := Controller{Api: api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: true,
		MaxBackoff:    30,
		MinBackoff:    1,
		MaxRetries:    4,
	})}
	ctx := context.Background()

	// The CI VM only has a single NIC, so assign a VLAN device on top of it.
	vlanID, err := c.AddVlan(ctx, &Vlan{Description: "assign test", Parent: "vtnet0", Tag: "3901", Priority: "0"})
	if err != nil {
		t.Fatalf("AddVlan: %v", err)
	}
	vlan, err := c.GetVlan(ctx, vlanID)
	if err != nil {
		t.Fatalf("GetVlan: %v", err)
	}
	t.Cleanup(func() {
		if err := c.DeleteVlan(ctx, vlanID); err != nil {
			t.Errorf("DeleteVlan: %v", err)
		}
	})

	a := &Assignment{
		Description: "TFTEST",
		Device:      api.SelectedMap(vlan.Device),
		Enable:      "1",
		Lock:        "0",
		IPv4Type:    "staticv4",
		IPv4Address: "192.0.2.1/24",
	}
	ifname, err := c.AddAssignment(ctx, a)
	if err != nil {
		t.Fatalf("AddAssignment: %v", err)
	}
	t.Logf("assigned %s as %s", vlan.Device, ifname)
	t.Cleanup(func() {
		if err := c.DeleteAssignment(ctx, ifname); err != nil {
			t.Errorf("DeleteAssignment: %v", err)
		}
	})

	got, err := c.GetAssignment(ctx, ifname)
	if err != nil {
		t.Fatalf("GetAssignment: %v", err)
	}
	if got.Description != a.Description || string(got.Device) != vlan.Device || got.IPv4Address != a.IPv4Address {
		t.Fatalf("GetAssignment mismatch: %+v", got)
	}

	a.Description = "TFTEST2"
	a.IPv4Address = "192.0.2.2/24"
	if err := c.UpdateAssignment(ctx, ifname, a); err != nil {
		t.Fatalf("UpdateAssignment: %v", err)
	}
	got, err = c.GetAssignment(ctx, ifname)
	if err != nil {
		t.Fatalf("GetAssignment after update: %v", err)
	}
	if got.Description != "TFTEST2" || got.IPv4Address != "192.0.2.2/24" {
		t.Fatalf("update not applied: %+v", got)
	}
}
