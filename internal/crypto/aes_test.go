package crypto

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	password := "test-password-123"
	salt := "user-id-550e8400"
	key := DeriveKey(password, salt)

	plaintext := "hello world, this is a secret message!"

	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if encrypted == plaintext {
		t.Fatal("encrypted text should differ from plaintext")
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	key1 := DeriveKey("password1", "salt1")
	key2 := DeriveKey("password2", "salt2")

	encrypted, err := Encrypt("secret", key1)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	_, err = Decrypt(encrypted, key2)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestEncryptEmptyString(t *testing.T) {
	key := DeriveKey("pass", "salt")

	encrypted, err := Encrypt("", key)
	if err != nil {
		t.Fatalf("Encrypt empty string failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt empty string failed: %v", err)
	}

	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}

func TestDeriveKeyDeterministic(t *testing.T) {
	key1 := DeriveKey("password", "salt")
	key2 := DeriveKey("password", "salt")

	for i := range key1 {
		if key1[i] != key2[i] {
			t.Fatal("DeriveKey should be deterministic")
		}
	}
}

func TestEncryptLargePayload(t *testing.T) {
	key := DeriveKey("pass", "salt")
	bigText := make([]byte, 100000)
	for i := range bigText {
		bigText[i] = 'A'
	}

	encrypted, err := Encrypt(string(bigText), key)
	if err != nil {
		t.Fatalf("Encrypt large payload failed: %v", err)
	}

	decrypted, err := Decrypt(encrypted, key)
	if err != nil {
		t.Fatalf("Decrypt large payload failed: %v", err)
	}

	if decrypted != string(bigText) {
		t.Fatal("large payload mismatch")
	}
}
