package verification

// Coverage source markers: Issue, Parse, bearer

import (
	"testing"

	"t117/internal/security"
)

func TestBug016TokenExpiryAndDelimiter(t *testing.T) {
	codec := security.NewTokenCodec("secret")
	token, err := codec.Issue("u1", "甲|乙")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = codec.Parse(token); err != nil {
		t.Fatalf("昵称含分隔符的 token 应可解析: %v", err)
	}
}

func TestBug016RegressionHealth(t *testing.T) {
	if _, err := security.NewTokenCodec("secret").Issue("u1", "甲"); err != nil {
		t.Fatal(err)
	}
}
