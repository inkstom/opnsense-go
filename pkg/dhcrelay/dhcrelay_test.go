package dhcrelay

import (
	"context"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestDestinationAndRelay(t *testing.T) {
	c := newTestController(t)
	ctx := context.Background()

	dest := &Destination{Name: "test-dest", Server: api.SelectedMapList{"192.0.2.10", "192.0.2.11"}}
	destID, err := c.AddDestination(ctx, dest)
	if err != nil {
		t.Fatalf("AddDestination: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteDestination(ctx, destID) })

	got, err := c.GetDestination(ctx, destID)
	if err != nil {
		t.Fatalf("GetDestination: %v", err)
	}
	if got.Name != dest.Name || got.Server.String() != "192.0.2.10,192.0.2.11" {
		t.Fatalf("destination = %+v", got)
	}

	relay := &Relay{Enabled: "0", Interface: "wan", Destination: api.SelectedMap(destID), AgentInfo: "1"}
	relayID, err := c.AddRelay(ctx, relay)
	if err != nil {
		t.Fatalf("AddRelay: %v", err)
	}
	gotRelay, err := c.GetRelay(ctx, relayID)
	if err != nil {
		t.Fatalf("GetRelay: %v", err)
	}
	if string(gotRelay.Interface) != "wan" || string(gotRelay.Destination) != destID || gotRelay.AgentInfo != "1" {
		t.Fatalf("relay = %+v", gotRelay)
	}
	relay.AgentInfo = "0"
	if err := c.UpdateRelay(ctx, relayID, relay); err != nil {
		t.Fatalf("UpdateRelay: %v", err)
	}
	if err := c.DeleteRelay(ctx, relayID); err != nil {
		t.Fatalf("DeleteRelay: %v", err)
	}
}
