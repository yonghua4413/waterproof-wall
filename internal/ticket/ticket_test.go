package ticket

import (
	"testing"
	"time"
)

func TestIssueAndParse(t *testing.T) {
	secret := []byte("test-secret-32-bytes-long-enough")
	token, payload, err := Issue(secret, "app_test", "captcha-id", "login", "user-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := Parse(secret, token)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.TicketID != payload.TicketID || parsed.CaptchaID != "captcha-id" {
		t.Fatalf("unexpected payload: %#v", parsed)
	}
}

func TestParseRejectsTamperedTicket(t *testing.T) {
	secret := []byte("test-secret-32-bytes-long-enough")
	token, _, err := Issue(secret, "app_test", "captcha-id", "login", "user-1", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Parse(secret, token+"x")
	if err != ErrSignature {
		t.Fatalf("expected ErrSignature, got %v", err)
	}
}
