package auth

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestMain은 패키지의 모든 테스트 전에 한 번 실행된다.
func TestMain(m *testing.M) {
	Cost = bcrypt.MinCost
	m.Run()
}

func TestPasswordRoundTrip(t *testing.T) {
	h, err := HashPassword("correct horse")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(h, "correct horse") {
		t.Error("correct password rejected")
	}
	if CheckPassword(h, "wrong horse") {
		t.Error("wrong password accepted")
	}
	if CheckPassword("", "anything") {
		t.Error("empty hash (unknown user) must never match")
	}

	h2, _ := HashPassword("correct horse")
	if h == h2 {
		t.Error("hashes of same password should differ (salt)")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		in string
		ok bool
	}{
		{"short", false},
		{"12345678", true},
		{strings.Repeat("a", 72), true},
		{strings.Repeat("a", 73), false},
	}
	for _, tt := range tests {
		if err := ValidatePassword(tt.in); (err == nil) != tt.ok {
			t.Errorf("ValidatePassword(len=%d) err=%v, want ok=%v", len(tt.in), err, tt.ok)
		}
	}
}

func TestToken(t *testing.T) {
	a, b := NewToken(), NewToken()
	if a == b {
		t.Error("tokens must be unique")
	}
	if len(a) != 43 { // 32바이트 → base64 raw 43자
		t.Errorf("token length = %d", len(a))
	}
	if HashToken(a) != HashToken(a) || HashToken(a) == HashToken(b) {
		t.Error("HashToken must be deterministic and distinct")
	}
}
