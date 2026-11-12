package firewall

import (
	"context"
	"os"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestGroup(t *testing.T) {
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

	id, err := c.AddGroup(ctx, &Group{Name: "testgrp", Members: api.SelectedMapList{"wan", "lo0"}, NoGroup: "1",
		Sequence: "5", Description: "test group"})
	if err != nil {
		t.Fatalf("AddGroup: %v", err)
	}
	g, err := c.GetGroup(ctx, id)
	if err != nil || g.Name != "testgrp" || len(g.Members) != 2 || g.NoGroup != "1" {
		t.Fatalf("GetGroup = %+v, %v", g, err)
	}
	g.Members = api.SelectedMapList{"wan"}
	g.Description = "updated"
	if err := c.UpdateGroup(ctx, id, g); err != nil {
		t.Fatalf("UpdateGroup: %v", err)
	}
	g, err = c.GetGroup(ctx, id)
	if err != nil || len(g.Members) != 1 || g.Description != "updated" {
		t.Fatalf("GetGroup after update = %+v, %v", g, err)
	}
	if err := c.DeleteGroup(ctx, id); err != nil {
		t.Fatalf("DeleteGroup: %v", err)
	}
}
