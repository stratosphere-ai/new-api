package common

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func GenerateHMACWithKey(key []byte, data string) string {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func GenerateHMAC(data string) string {
	h := hmac.New(sha256.New, []byte(CryptoSecret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func Password2Hash(password string) (string, error) {
	passwordBytes := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	return string(hashedPassword), err
}

func ValidatePasswordAndHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// MarketplaceEncryptionKey is loaded from MARKETPLACE_ENCRYPTION_KEY env var.
// Must be exactly 32 bytes for AES-256.
var MarketplaceEncryptionKey []byte

func InitMarketplaceEncryption() {
	keyStr := os.Getenv("MARKETPLACE_ENCRYPTION_KEY")
	if keyStr == "" {
		return
	}
	// Derive a 32-byte key from the env var using SHA-256
	hash := sha256.Sum256([]byte(keyStr))
	MarketplaceEncryptionKey = hash[:]
}

// EncryptAPIKey encrypts an API key using AES-256-GCM.
// Returns "enc:" + base64(nonce + ciphertext).
func EncryptAPIKey(plainKey string) (string, error) {
	if len(MarketplaceEncryptionKey) == 0 {
		return "", errors.New("marketplace encryption key not configured")
	}
	block, err := aes.NewCipher(MarketplaceEncryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainKey), nil)
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptAPIKey decrypts an "enc:"-prefixed encrypted API key.
// If the key doesn't have the prefix, it's returned as-is (backward compat).
func DecryptAPIKey(encryptedKey string) (string, error) {
	if !strings.HasPrefix(encryptedKey, "enc:") {
		return encryptedKey, nil
	}
	if len(MarketplaceEncryptionKey) == 0 {
		return "", errors.New("marketplace encryption key not configured")
	}
	data, err := base64.StdEncoding.DecodeString(encryptedKey[4:])
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(MarketplaceEncryptionKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
