package auth

import "testing"

func TestValidateToken(t *testing.T) {
	t.Parallel()
	s := Load(t.TempDir())
	acct, ok := s.ValidateToken("demo-token-change-me")
	if !ok || acct.ID != "demo" {
		t.Fatalf("expected demo account, got %+v ok=%v", acct, ok)
	}
	if _, ok := s.ValidateToken("bad"); ok {
		t.Fatal("bad token should fail")
	}
}

func TestExtractToken(t *testing.T) {
	t.Parallel()
	if got := ExtractToken("Bearer abc", "", ""); got != "abc" {
		t.Fatalf("want abc got %q", got)
	}
	if got := ExtractToken("", "qtok", ""); got != "qtok" {
		t.Fatalf("want qtok got %q", got)
	}
}
