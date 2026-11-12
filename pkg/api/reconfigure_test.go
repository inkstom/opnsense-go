package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestReconfigureService(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"ok", `{"status":"ok"}`, false},
		{"wireguard result", `{"result":"ok"}`, false},
		{"disabled service already stopped (26.7.6+)", `{"status":"failed","status_msg":"Service \"ipsec\" failed to stop."}`, false},
		{"start failure", `{"status":"failed","status_msg":"Service \"ipsec\" failed to start."}`, true},
		{"reload failure", `{"status":"failed","status_msg":"Service \"unbound\" failed to reload."}`, true},
		{"plain failure", `{"status":"failed"}`, true},
		{"unknown", `{}`, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			c := NewClient(Options{Uri: srv.URL, APIKey: "k", APISecret: "s", MaxRetries: 1, MinBackoff: 1, MaxBackoff: 1})
			err := c.ReconfigureService(context.Background(), "/test/service/reconfigure")
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
