package core

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestHasync(t *testing.T) {
	if os.Getenv("OPNSENSE_URI") == "" {
		t.Skip("OPNSENSE_URI not set")
	}
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

	before, err := c.HasyncGet(ctx)
	if err != nil {
		t.Fatalf("HasyncGet: %v", err)
	}
	t.Cleanup(func() {
		if _, err := c.HasyncSet(ctx, &before.Hasync); err != nil {
			t.Errorf("restore: %v", err)
		}
		_, _ = c.HasyncReconfigure(ctx)
	})

	want := before.Hasync
	want.SynchronizeToIp = "192.0.2.10"
	want.Username = "root"
	want.Password = "test-secret"
	want.SyncItems = api.SelectedMapList{"aliases", "cron"}
	res, err := c.HasyncSet(ctx, &want)
	if err != nil || res.Result != "saved" {
		t.Fatalf("HasyncSet = %+v, %v", res, err)
	}
	if _, err := c.HasyncReconfigure(ctx); err != nil {
		t.Fatalf("HasyncReconfigure: %v", err)
	}
	got, err := c.HasyncGet(ctx)
	if err != nil {
		t.Fatalf("HasyncGet: %v", err)
	}
	h := got.Hasync
	if h.SynchronizeToIp != "192.0.2.10" || h.Password != "test-secret" || len(h.SyncItems) != 2 || h.PfsyncVersion == "" {
		t.Fatalf("HasyncGet after set = %+v", h)
	}
}
