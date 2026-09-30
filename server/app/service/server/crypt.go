package server

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/chacha20poly1305"
)

// encrypt 加密，失败时返回空字符串。
func (s *service) encrypt(key, data string) string {
	if data == "" {
		return ""
	}
	innerKey := sha256.Sum256([]byte("terminal-xchacha:" + key))
	inner, err := chacha20poly1305.NewX(innerKey[:])
	if err != nil {
		return ""
	}
	innerNonce := make([]byte, inner.NonceSize())
	if _, err = io.ReadFull(rand.Reader, innerNonce); err != nil {
		return ""
	}
	innerData := inner.Seal(innerNonce, innerNonce, []byte(data), []byte(key))
	outerKey := sha256.Sum256([]byte("terminal-aes:" + key))
	block, err := aes.NewCipher(outerKey[:])
	if err != nil {
		return ""
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return ""
	}
	sealed := aead.Seal(nonce, nonce, innerData, []byte(key))
	return base64.RawURLEncoding.EncodeToString(sealed)
}

// decrypt 解密，失败时返回空字符串。
func (s *service) decrypt(key, data string) string {
	if data == "" {
		return ""
	}
	ciphertext, err := base64.RawURLEncoding.DecodeString(data)
	if err != nil {
		return ""
	}
	outerKey := sha256.Sum256([]byte("terminal-aes:" + key))
	block, err := aes.NewCipher(outerKey[:])
	if err != nil {
		return ""
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return ""
	}
	if len(ciphertext) < aead.NonceSize() {
		return ""
	}
	innerData, err := aead.Open(nil, ciphertext[:aead.NonceSize()], ciphertext[aead.NonceSize():], []byte(key))
	if err != nil {
		return ""
	}
	innerKey := sha256.Sum256([]byte("terminal-xchacha:" + key))
	inner, err := chacha20poly1305.NewX(innerKey[:])
	if err != nil {
		return ""
	}
	if len(innerData) < inner.NonceSize() {
		return ""
	}
	plain, err := inner.Open(nil, innerData[:inner.NonceSize()], innerData[inner.NonceSize():], []byte(key))
	if err != nil {
		return ""
	}
	return string(plain)
}
