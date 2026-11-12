package trafficshaper

import (
	"context"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestPipeQueueRule(t *testing.T) {
	c := newTestController(t)
	ctx := context.Background()

	pipeID, err := c.AddPipe(ctx, &Pipe{Enabled: "1", Bandwidth: "50", BandwidthMetric: "Mbit", Mask: "none",
		Scheduler: "fq_codel", Description: "test pipe"})
	if err != nil {
		t.Fatalf("AddPipe: %v", err)
	}
	t.Cleanup(func() { _ = c.DeletePipe(ctx, pipeID) })
	pipe, err := c.GetPipe(ctx, pipeID)
	if err != nil || pipe.Bandwidth != "50" || pipe.Scheduler != "fq_codel" || pipe.Number == "" {
		t.Fatalf("GetPipe = %+v, %v", pipe, err)
	}

	queueID, err := c.AddQueue(ctx, &Queue{Enabled: "1", Pipe: api.SelectedMap(pipeID), Weight: "50", Mask: "none",
		Description: "test queue"})
	if err != nil {
		t.Fatalf("AddQueue: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteQueue(ctx, queueID) })
	queue, err := c.GetQueue(ctx, queueID)
	if err != nil || queue.Pipe.String() != pipeID || queue.Weight != "50" {
		t.Fatalf("GetQueue = %+v, %v", queue, err)
	}

	ruleID, err := c.AddRule(ctx, &Rule{Enabled: "1", Sequence: "10", Interface: "wan", Protocol: "tcp",
		Source: api.SelectedMapList{"192.0.2.0/24", "198.51.100.0/24"}, SourceNot: "0", SourcePort: "any",
		Destination: api.SelectedMapList{"any"}, DestinationNot: "0", DestinationPort: "443",
		Dscp: api.SelectedMapList{"ef"}, Direction: "out", Target: api.SelectedMap(queueID), Description: "test rule"})
	if err != nil {
		t.Fatalf("AddRule: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteRule(ctx, ruleID) })
	rule, err := c.GetRule(ctx, ruleID)
	if err != nil || rule.Target.String() != queueID || len(rule.Source) != 2 || rule.DestinationPort != "443" {
		t.Fatalf("GetRule = %+v, %v", rule, err)
	}

	rule.Target = api.SelectedMap(pipeID)
	rule.Dscp = api.SelectedMapList{}
	if err := c.UpdateRule(ctx, ruleID, rule); err != nil {
		t.Fatalf("UpdateRule: %v", err)
	}
	rule, err = c.GetRule(ctx, ruleID)
	if err != nil || rule.Target.String() != pipeID || len(rule.Dscp) != 0 {
		t.Fatalf("GetRule after update = %+v, %v", rule, err)
	}

	for _, del := range []func() error{
		func() error { return c.DeleteRule(ctx, ruleID) },
		func() error { return c.DeleteQueue(ctx, queueID) },
		func() error { return c.DeletePipe(ctx, pipeID) },
	} {
		if err := del(); err != nil {
			t.Fatalf("delete: %v", err)
		}
	}
}
