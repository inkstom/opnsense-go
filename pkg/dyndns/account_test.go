package dyndns

import (
	"context"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func TestAccount(t *testing.T) {
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

	// Disabled, so ddclient never contacts example.com.
	acct := &Account{
		Enabled:        "0",
		Service:        "custom",
		Protocol:       "dyndns2",
		Server:         "dyndns.example.com",
		Username:       "tf-test",
		Password:       "s3cret",
		Hostnames:      api.SelectedMapList{"tf-test.example.com", "tf-test2.example.com"},
		Wildcard:       "0",
		Checkip:        "web_icanhazip",
		CheckipTimeout: "10",
		ForceSsl:       "1",
		Ttl:            "300",
		Description:    "tf-test-dyndns",
	}
	id, err := c.AddAccount(ctx, acct)
	if err != nil {
		if strings.Contains(err.Error(), "status code 404") {
			t.Skip("os-ddclient plugin not installed")
		}
		t.Fatalf("AddAccount: %v", err)
	}
	t.Cleanup(func() { _ = c.DeleteAccount(ctx, id) })

	got, err := c.GetAccount(ctx, id)
	if err != nil {
		t.Fatalf("GetAccount: %v", err)
	}
	// The password is write-only: GET always returns it empty.
	if got.Service != "custom" || got.Protocol != "dyndns2" || got.Password != "" ||
		!slices.Equal(got.Hostnames, acct.Hostnames) || got.Checkip != "web_icanhazip" {
		t.Fatalf("account = %+v", got)
	}

	acct.Hostnames = api.SelectedMapList{"tf-test3.example.com"}
	acct.Password = "" // empty keeps the stored password
	acct.Ttl = "600"
	if err := c.UpdateAccount(ctx, id, acct); err != nil {
		t.Fatalf("UpdateAccount: %v", err)
	}
	if got, _ = c.GetAccount(ctx, id); !slices.Equal(got.Hostnames, acct.Hostnames) || got.Ttl != "600" {
		t.Fatalf("account after update = %+v", got)
	}
	if err := c.DeleteAccount(ctx, id); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}
}
