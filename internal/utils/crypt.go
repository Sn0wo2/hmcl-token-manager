package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/ascii85"
	"errors"
	"fmt"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	"golang.org/x/crypto/scrypt"
)

const (
	// 别改长度
	magicHead = "HMCLTM3"

	saltSize  = 16
	keySize   = 16
	nonceSize = 12
	tagSize   = 12
)

func Encrypt[T any](v T, password string) (string, error) {
	plain, err := msgpack.Marshal(v)
	if err != nil {
		return "", err
	}

	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	material, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, keySize+nonceSize)
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

	payload := make([]byte, len(magicHead)+saltSize, len(magicHead)+saltSize+len(plain)+tagSize)
	copy(payload, magicHead)
	copy(payload[len(magicHead):], salt)
	payload = gcm.Seal(payload, material[keySize:], plain, nil)
	encoded := make([]byte, ascii85.MaxEncodedLen(len(payload)))
	return string(encoded[:ascii85.Encode(encoded, payload)]), nil
}

func Decrypt[T any](encoded, password string) (T, error) {
	var zero T

	payload := make([]byte, len(encoded)+3*strings.Count(encoded, "z"))
	n, _, err := ascii85.Decode(payload, []byte(encoded), true)
	if err != nil {
		return zero, err
	}
	payload = payload[:n]

	if len(payload) < len(magicHead)+saltSize+tagSize {
		return zero, errors.New("invalid token")
	}

	if string(payload[:len(magicHead)]) != magicHead {
		return zero, fmt.Errorf("unsupported magic head %s, expected %s", string(payload[:len(magicHead)]), magicHead)
	}

	salt := payload[len(magicHead) : len(magicHead)+saltSize]
	material, err := scrypt.Key([]byte(password), salt, 1<<15, 8, 1, keySize+nonceSize)
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

	plain, err := gcm.Open(nil, material[keySize:], payload[len(magicHead)+saltSize:], nil)
	if err != nil {
		return zero, err
	}

	var v T
	if err := msgpack.Unmarshal(plain, &v); err != nil {
		return zero, err
	}

	return v, nil
}
