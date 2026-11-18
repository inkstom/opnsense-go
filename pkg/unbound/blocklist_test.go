package unbound

import (
	"context"
	"os"
	"slices"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestBlocklist(t *testing.T) {
	c := &Controller{Api: api.NewClient(api.Options{
		Uri:           os.Getenv("OPNSENSE_URI"),
		APIKey:        os.Getenv("OPNSENSE_API_KEY"),
		APISecret:     os.Getenv("OPNSENSE_API_SECRET"),
		AllowInsecure: true,
		MaxBackoff:    30,
		MinBackoff:    1,
		MaxRetries:    4,
	})}
	ctx := context.Background()

	// Only local lists, so the apply doesn't download anything.
	bl := &Blocklist{
		Enabled:     "1",
		Blocklists:  api.SelectedMapList{"bad.example.com", "worse.example.org"},
		Allowlists:  api.SelectedMapList{"good.example.com"},
		Wildcards:   api.SelectedMapList{"ads.example.net"},
		SourceNets:  api.SelectedMapList{"192.0.2.0/24", "198.51.100.0/24"},
		Address:     "192.0.2.1",
		NXDomain:    "0",
		CacheTTL:    "3600",
		Description: "tf-test-dnsbl",
	}
	id, err := c.AddBlocklist(ctx, bl)
	if err != nil {
		t.Fatalf("AddBlocklist: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteBlocklist(ctx, id) })

	got, err := c.GetBlocklist(ctx, id)
	if err != nil {
		t.Fatalf("GetBlocklist: %v", err)
	}
	if !slices.Equal(got.Blocklists, bl.Blocklists) || !slices.Equal(got.SourceNets, bl.SourceNets) ||
		len(got.Type) != 0 || len(got.Lists) != 0 || got.CacheTTL != "3600" || got.Address != "192.0.2.1" {
		t.Fatalf("blocklist = %+v", got)
	}

	bl.Enabled = "0"
	bl.Type = api.SelectedMapList{"hgz002", "oisd0"}
	bl.SourceNets = api.SelectedMapList{}
	bl.NXDomain = "1"
	if err := c.UpdateBlocklist(ctx, id, bl); err != nil {
		t.Fatalf("UpdateBlocklist: %v", err)
	}
	if got, _ = c.GetBlocklist(ctx, id); !slices.Equal(got.Type, bl.Type) || len(got.SourceNets) != 0 || got.NXDomain != "1" {
		t.Fatalf("blocklist after update = %+v", got)
	}
	if err := c.DeleteBlocklist(ctx, id); err != nil {
		t.Fatalf("DeleteBlocklist: %v", err)
	}
}
