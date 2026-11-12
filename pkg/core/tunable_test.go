package core

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestTunable(t *testing.T) {
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

	tun := &Tunable{Tunable: "net.inet.tcp.blackhole", Value: "2", Description: "test tunable"}
	id, err := c.AddTunable(ctx, tun)
	if err != nil {
		t.Fatalf("AddTunable: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteTunable(ctx, id) })

	got, err := c.GetTunable(ctx, id)
	if err != nil {
		t.Fatalf("GetTunable: %v", err)
	}
	if got.Tunable != tun.Tunable || got.Value != "2" || got.Description != tun.Description {
		t.Fatalf("tunable = %+v", got)
	}
	tun.Value = "1"
	if err := c.UpdateTunable(ctx, id, tun); err != nil {
		t.Fatalf("UpdateTunable: %v", err)
	}
	if got, _ = c.GetTunable(ctx, id); got.Value != "1" {
		t.Fatalf("value after update = %q", got.Value)
	}
	if err := c.DeleteTunable(ctx, id); err != nil {
		t.Fatalf("DeleteTunable: %v", err)
	}
}
