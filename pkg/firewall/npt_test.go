package firewall

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestNPT(t *testing.T) {
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

	id, err := c.AddNPT(ctx, &NPT{Enabled: "0", Log: "1", Sequence: "10", Interface: "wan",
		SourceNet: "fd00:1::/64", DestinationNet: "2001:db8:1::/64", Description: "test npt"})
	if err != nil {
		t.Fatalf("AddNPT: %v", err)
	}
	r, err := c.GetNPT(ctx, id)
	if err != nil || r.Interface != "wan" || r.DestinationNet != "2001:db8:1::/64" || len(r.Categories) != 0 {
		t.Fatalf("GetNPT = %+v, %v", r, err)
	}
	r.DestinationNet = "2001:db8:2::/64"
	if err := c.UpdateNPT(ctx, id, r); err != nil {
		t.Fatalf("UpdateNPT: %v", err)
	}
	if r, err = c.GetNPT(ctx, id); err != nil || r.DestinationNet != "2001:db8:2::/64" {
		t.Fatalf("GetNPT after update = %+v, %v", r, err)
	}
	if err := c.DeleteNPT(ctx, id); err != nil {
		t.Fatalf("DeleteNPT: %v", err)
	}
}
