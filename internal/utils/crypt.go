package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Sn0wo2/hmcl-token-manager/internal/account"
	"golang.org/x/crypto/scrypt"
)

func Encrypt(acc account.Account, password string) (string, error) {
	plain, err := json.Marshal(acc)
	if err != nil {
		return "", fmt.Errorf("marshal account: %w", err)
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}

	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return "", fmt.Errorf("derive key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("new cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("new gcm: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, plain, nil)

	payload := make([]byte, 0, 7+len(salt)+len(nonce)+len(ciphertext))
	payload = append(payload, []byte("HMCLTM1")...)
	payload = append(payload, salt...)
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func Decrypt(encoded, password string) (account.Account, error) {
	var acc account.Account

	payload, err := base64.RawURLEncoding.DecodeString(strings.TrimSpace(encoded))
	if err != nil {
		return acc, fmt.Errorf("invalid encoded data: %w", err)
	}

	if len(payload) < 7+16+12+16 || !bytes.Equal(payload[:7], []byte("HMCLTM1")) {
		return acc, fmt.Errorf("invalid HMCL token data")
	}

	salt := payload[7:23]
	nonce := payload[23:35]
	ciphertext := payload[35:]

	key, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, 32)
	if err != nil {
		return acc, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return acc, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return acc, err
	}

	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return acc, fmt.Errorf("wrong password or corrupted data")
	}

	if err := json.Unmarshal(plain, &acc); err != nil {
		return acc, fmt.Errorf("parse account data: %w", err)
	}

	return acc, nil
}
