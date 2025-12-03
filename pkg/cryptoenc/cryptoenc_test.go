package cryptoenc

import (
	"testing"
)

type testStruct struct {
	Foo string `json:"foo"`
	Bar int    `json:"bar"`
}

func TestEncryptDecryptToString_RoundTrip(t *testing.T) {
	key := "test-key-123"
	plain := "hello secret world"

	enc, err := EncryptToString(plain, key)
	if err != nil {
		t.Fatalf("EncryptToString error: %v", err)
	}
	if enc == "" {
		t.Fatal("expected non-empty ciphertext")
	}

	got, err := DecryptFromString(enc, key)
	if err != nil {
		t.Fatalf("DecryptFromString error: %v", err)
	}
	if got != plain {
		t.Fatalf("got %q, want %q", got, plain)
	}
}

func TestEncryptDecryptToString_WrongKey(t *testing.T) {
	key := "test-key-123"
	badKey := "other-key-456"
	plain := "hello secret world"

	enc, err := EncryptToString(plain, key)
	if err != nil {
		t.Fatalf("EncryptToString error: %v", err)
	}

	if _, err := DecryptFromString(enc, badKey); err == nil {
		t.Fatal("expected error on decrypt with wrong key, got nil")
	}
}

func TestEncryptDecryptJSON_RoundTrip(t *testing.T) {
	key := "json-key"
	in := testStruct{Foo: "abc", Bar: 42}

	enc, err := EncryptJSON(in, key)
	if err != nil {
		t.Fatalf("EncryptJSON error: %v", err)
	}
	if enc == "" {
		t.Fatal("expected non-empty ciphertext from EncryptJSON")
	}

	var out testStruct
	if err := DecryptJSON(enc, key, &out); err != nil {
		t.Fatalf("DecryptJSON error: %v", err)
	}

	if out.Foo != in.Foo || out.Bar != in.Bar {
		t.Fatalf("round-trip mismatch: got %+v, want %+v", out, in)
	}
}

func TestDecryptJSON_InvalidBase64(t *testing.T) {
	key := "json-key"
	// заведомо невалидная base64-строка
	err := DecryptJSON("!!!not-base64!!!", key, &testStruct{})
	if err == nil {
		t.Fatal("expected error on invalid base64, got nil")
	}
}
