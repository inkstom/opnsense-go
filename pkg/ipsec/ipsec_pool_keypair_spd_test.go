package ipsec

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"os"
	"strings"
	"testing"

	"github.com/browningluke/opnsense-go/pkg/api"
)

func newPoolTestController() *Controller {
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

func TestIPsecPool(t *testing.T) {
	c := newPoolTestController()
	ctx := context.Background()
	id, err := c.AddIPsecPool(ctx, &IPsecPool{Enabled: "1", Name: "testpool", Addresses: "10.200.0.0/24",
		Dns: api.SelectedMapList{"192.0.2.53", "192.0.2.54"}})
	if err != nil {
		t.Fatalf("AddIPsecPool: %v", err)
	}
	p, err := c.GetIPsecPool(ctx, id)
	if err != nil || p.Addresses != "10.200.0.0/24" || len(p.Dns) != 2 {
		t.Fatalf("GetIPsecPool = %+v, %v", p, err)
	}
	p.Dns = api.SelectedMapList{}
	if err := c.UpdateIPsecPool(ctx, id, p); err != nil {
		t.Fatalf("UpdateIPsecPool: %v", err)
	}
	if p, err = c.GetIPsecPool(ctx, id); err != nil || len(p.Dns) != 0 {
		t.Fatalf("GetIPsecPool after update = %+v, %v", p, err)
	}
	if err := c.DeleteIPsecPool(ctx, id); err != nil {
		t.Fatalf("DeleteIPsecPool: %v", err)
	}
}

func TestIPsecKeyPair(t *testing.T) {
	c := newPoolTestController()
	ctx := context.Background()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	pubDER, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
	privDER, _ := x509.MarshalPKCS8PrivateKey(key)
	pub := string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubDER}))
	priv := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privDER}))

	id, err := c.AddIPsecKeyPair(ctx, &IPsecKeyPair{Name: "testkp", KeyType: "ecdsa", PublicKey: pub, PrivateKey: priv})
	if err != nil {
		t.Fatalf("AddIPsecKeyPair: %v", err)
	}
	kp, err := c.GetIPsecKeyPair(ctx, id)
	if err != nil || kp.KeySize != "256" || kp.KeyFingerprint == "" || strings.TrimSpace(kp.PublicKey) != strings.TrimSpace(pub) {
		t.Fatalf("GetIPsecKeyPair = %+v, %v", kp, err)
	}
	if err := c.DeleteIPsecKeyPair(ctx, id); err != nil {
		t.Fatalf("DeleteIPsecKeyPair: %v", err)
	}
}

func TestIPsecManualSPD(t *testing.T) {
	c := newPoolTestController()
	ctx := context.Background()
	id, err := c.AddIPsecManualSPD(ctx, &IPsecManualSPD{Enabled: "1", Protocol: "esp", ReqId: "100",
		Source: "10.0.0.0/24", Destination: "10.1.0.0/24", Description: "test spd"})
	if err != nil {
		t.Fatalf("AddIPsecManualSPD: %v", err)
	}
	s, err := c.GetIPsecManualSPD(ctx, id)
	if err != nil || s.ReqId != "100" || s.Protocol != "esp" || s.Destination != "10.1.0.0/24" {
		t.Fatalf("GetIPsecManualSPD = %+v, %v", s, err)
	}
	if err := c.DeleteIPsecManualSPD(ctx, id); err != nil {
		t.Fatalf("DeleteIPsecManualSPD: %v", err)
	}
}
