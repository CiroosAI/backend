package middleware

import (
	"net/http/httptest"
	"testing"
)

func TestClientIPGeneric_DirectRemote(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.local/", nil)
	req.RemoteAddr = "203.0.113.5:54321"
	ip := clientIPGeneric(req, nil)
	if ip != "203.0.113.5" {
		t.Fatalf("expected direct remote IP, got %s", ip)
	}
}

func TestClientIPGeneric_TrustedProxyXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.local/", nil)
	req.RemoteAddr = "198.51.100.10:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.7, 198.51.100.10")
	// trustedCIDR contains the remote IP
	ip := clientIPGeneric(req, []string{"198.51.100.10"})
	if ip != "203.0.113.7" {
		t.Fatalf("expected X-Forwarded-For first value, got %s", ip)
	}
}

func TestClientIPGeneric_UntrustedProxyIgnoresXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "http://example.local/", nil)
	req.RemoteAddr = "198.51.100.11:443"
	req.Header.Set("X-Forwarded-For", "203.0.113.8, 198.51.100.11")
	ip := clientIPGeneric(req, []string{"198.51.100.10"})
	if ip != "198.51.100.11" {
		t.Fatalf("expected remote IP when proxy untrusted, got %s", ip)
	}
}
