package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/scrypt"
)

const (
	// 别改长度
	magicHead = "HMCLTM2"

	saltSize  = 16
	keySize   = 16
	nonceSize = 12
	tagSize   = 12
)

type Crypto struct{}

func (Crypto) Encrypt[T any](v T, password string) (string, error) {
	plain, err := msgpack.Marshal(v)
	if err != nil {
		return "", err
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	material, err := scrypt.Key(
		[]byte(password),
		salt,
		1<<15,
		8,
		1,
		keySize+nonceSize,
	)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(material[:keySize])
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(
		nil,
		material[keySize:],
		plain,
		nil,
	)

	payload := make([]byte, len(magicHead)+saltSize+len(ciphertext))

	offset := 0

	copy(payload[offset:], magicHead)
	offset += len(magicHead)

	copy(payload[offset:], salt)
	offset += saltSize

	copy(payload[offset:], ciphertext)

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func (Crypto) Decrypt[T any](encoded, password string) (T, error) {
	var zero T

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return zero, err
	}

	if len(payload) < len(magicHead)+saltSize+tagSize {
		return zero, errors.New("invalid token")
	}

	if string(payload[:len(magicHead)]) != magicHead {
		return zero, fmt.Errorf("unsupported magic head %s, expected %s", string(payload[:len(magicHead)]), magicHead)
	}

	offset := len(magicHead)

	salt := payload[offset : offset+saltSize]
	offset += saltSize

	ciphertext := payload[offset:]

	material, err := scrypt.Key(
		[]byte(password),
		salt,
		1<<15,
		8,
		1,
		keySize+nonceSize,
	)
	if err != nil {
		return zero, err
	}

	block, err := aes.NewCipher(material[:keySize])
	if err != nil {
		return zero, err
	}

	gcm, err := cipher.NewGCMWithTagSize(block, tagSize)
	if err != nil {
		return zero, err
	}

	plain, err := gcm.Open(
		nil,
		material[keySize:],
		ciphertext,
		nil,
	)
	if err != nil {
		return zero, errors.New("wrong password or corrupted data")
	}

	var v T
	if err := msgpack.Unmarshal(plain, &v); err != nil {
		return zero, err
	}

	return v, nil
}