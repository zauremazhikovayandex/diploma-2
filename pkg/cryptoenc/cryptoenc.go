package cryptoenc

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// deriveKey — получает 32-байтный ключ из произвольной строки
func deriveKey(keyStr string) []byte {
	sum := sha256.Sum256([]byte(keyStr))
	return sum[:] // 32 байта
}

// EncryptToString шифрует plaintext и возвращает base64(nonce||ciphertext)
func EncryptToString(plaintext string, keyStr string) (string, error) {
	key := deriveKey(keyStr)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aead.Seal(nil, nonce, []byte(plaintext), nil)

	// сохраняем как nonce||cipher, кодируем в base64 для хранения в text
	buf := append(nonce, cipherText...)
	return base64.StdEncoding.EncodeToString(buf), nil
}

// DecryptFromString принимает base64(nonce||ciphertext) и ключ — возвращает plaintext
func DecryptFromString(data string, keyStr string) (string, error) {
	key := deriveKey(keyStr)

	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", fmt.Errorf("base64 decode: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aead.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce := raw[:nonceSize]
	cipherText := raw[nonceSize:]

	plain, err := aead.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt: %w", err)
	}
	return string(plain), nil
}

// EncryptJSON принимает любую структуру/мапу, маршалит в JSON и шифрует.
// Возвращает base64(nonce||ciphertext) как строку.
func EncryptJSON(v any, keyStr string) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return EncryptToString(string(data), keyStr)
}

// DecryptJSON расшифровывает base64(nonce||ciphertext), а затем делает json.Unmarshal в out.
// out должен быть указателем на структуру/мапу.
func DecryptJSON(enc string, keyStr string, out any) error {
	plain, err := DecryptFromString(enc, keyStr)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(plain), out)
}
