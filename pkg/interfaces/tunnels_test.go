package interfaces

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func tunnelsTestController() *Controller {
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

func TestLoopbackGifGreVxlanNeighbor(t *testing.T) {
	c := tunnelsTestController()
	ctx := context.Background()

	lo, err := c.AddLoopback(ctx, &Loopback{Description: "test loopback"})
	if err != nil {
		t.Fatalf("AddLoopback: %v", err)
	}
	gotLo, err := c.GetLoopback(ctx, lo)
	if err != nil || gotLo.Description != "test loopback" || gotLo.DeviceId == "" {
		t.Fatalf("GetLoopback = %+v, %v", gotLo, err)
	}
	if err := c.DeleteLoopback(ctx, lo); err != nil {
		t.Fatalf("DeleteLoopback: %v", err)
	}

	gif, err := c.AddGif(ctx, &Gif{LocalAddress: "wan", RemoteAddress: "192.0.2.1", TunnelLocalAddress: "10.99.0.1",
		TunnelRemoteAddress: "10.99.0.2", TunnelRemoteNet: "30", Description: "test gif", Link1: "0", Link2: "0", NoClamp: "0"})
	if err != nil {
		t.Fatalf("AddGif: %v", err)
	}
	gotGif, err := c.GetGif(ctx, gif)
	if err != nil || gotGif.RemoteAddress != "192.0.2.1" || gotGif.Device == "" {
		t.Fatalf("GetGif = %+v, %v", gotGif, err)
	}
	if err := c.DeleteGif(ctx, gif); err != nil {
		t.Fatalf("DeleteGif: %v", err)
	}

	gre, err := c.AddGre(ctx, &Gre{LocalAddress: "wan", RemoteAddress: "192.0.2.1", TunnelLocalAddress: "10.98.0.1",
		TunnelRemoteAddress: "10.98.0.2", TunnelRemoteNet: "30", Description: "test gre"})
	if err != nil {
		t.Fatalf("AddGre: %v", err)
	}
	gotGre, err := c.GetGre(ctx, gre)
	if err != nil || gotGre.TunnelLocalAddress != "10.98.0.1" {
		t.Fatalf("GetGre = %+v, %v", gotGre, err)
	}
	if err := c.DeleteGre(ctx, gre); err != nil {
		t.Fatalf("DeleteGre: %v", err)
	}

	vx, err := c.AddVxlan(ctx, &Vxlan{Vni: "100", LocalAddress: "10.0.2.15", RemoteAddress: "192.0.2.1"})
	if err != nil {
		t.Fatalf("AddVxlan: %v", err)
	}
	gotVx, err := c.GetVxlan(ctx, vx)
	if err != nil || gotVx.Vni != "100" {
		t.Fatalf("GetVxlan = %+v, %v", gotVx, err)
	}
	if err := c.DeleteVxlan(ctx, vx); err != nil {
		t.Fatalf("DeleteVxlan: %v", err)
	}

	nb, err := c.AddNeighbor(ctx, &Neighbor{MacAddress: "00:11:22:33:44:55", IpAddress: "10.0.2.200", Description: "test neighbor"})
	if err != nil {
		t.Fatalf("AddNeighbor: %v", err)
	}
	gotNb, err := c.GetNeighbor(ctx, nb)
	if err != nil || gotNb.MacAddress != "00:11:22:33:44:55" || gotNb.Origin != "manual" {
		t.Fatalf("GetNeighbor = %+v, %v", gotNb, err)
	}
	if err := c.DeleteNeighbor(ctx, nb); err != nil {
		t.Fatalf("DeleteNeighbor: %v", err)
	}
}
