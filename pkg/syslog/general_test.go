package syslog

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestGeneralSettings(t *testing.T) {
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

	before, err := c.GeneralGet(ctx)
	if err != nil {
		t.Fatalf("GeneralGet: %v", err)
	}
	t.Cleanup(func() {
		_, _ = c.GeneralSet(ctx, &before.Syslog)
		_, _ = c.GeneralReconfigure(ctx)
	})

	want := before.Syslog
	want.General.MaxPreserve = "20"
	want.General.MaxFileSize = "1000000"
	if res, err := c.GeneralSet(ctx, &want); err != nil || res.Result != "saved" {
		t.Fatalf("GeneralSet = %+v, %v", res, err)
	}
	if _, err := c.GeneralReconfigure(ctx); err != nil {
		t.Fatalf("GeneralReconfigure: %v", err)
	}
	got, err := c.GeneralGet(ctx)
	if err != nil || got.Syslog.General.MaxPreserve != "20" || got.Syslog.General.MaxFileSize != "1000000" || got.Syslog.General.Enabled != "1" {
		t.Fatalf("GeneralGet after set = %+v, %v", got, err)
	}
}
