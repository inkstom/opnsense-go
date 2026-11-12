package kea

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestCtrlAgent(t *testing.T) {
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

	before, err := c.CtrlAgentGet(ctx)
	if err != nil {
		t.Fatalf("CtrlAgentGet: %v", err)
	}
	t.Cleanup(func() { _, _ = c.CtrlAgentSet(ctx, &before.Ctrlagent) })

	want := before.Ctrlagent
	want.General.HttpPort = "8001"
	if res, err := c.CtrlAgentSet(ctx, &want); err != nil || res.Result != "saved" {
		t.Fatalf("CtrlAgentSet = %+v, %v", res, err)
	}
	got, err := c.CtrlAgentGet(ctx)
	if err != nil || got.Ctrlagent.General.HttpPort != "8001" || got.Ctrlagent.General.HttpHost == "" {
		t.Fatalf("CtrlAgentGet after set = %+v, %v", got, err)
	}
}
