package password

import "testing"

func TestOpaqueTokenIsRandomAndOnlyItsHashIsStable(t *testing.T) {
	first, err := GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("GenerateOpaqueToken() error = %v", err)
	}
	second, err := GenerateOpaqueToken()
	if err != nil {
		t.Fatalf("GenerateOpaqueToken() second error = %v", err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("opaque tokens must be independent high-entropy values: %q, %q", first, second)
	}
	firstHash := HashOpaqueToken(first)
	if string(firstHash) == first || len(firstHash) != 32 {
		t.Fatal("persisted token representation must be a SHA-256 hash, not the raw token")
	}
	if string(firstHash) != string(HashOpaqueToken(first)) || string(firstHash) == string(HashOpaqueToken(second)) {
		t.Fatal("token hash must be stable for one token and distinct for another")
	}
}
