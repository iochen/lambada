package utils

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

func cutKey(b []byte) []byte {
	l := len(b)
	switch {
	case l < 1:
		return []byte("no key provided!")
	case l < 5:
		return bytes.Repeat(b, 16)[0:16]
	case l < 8:
		return bytes.Repeat(b, 4)[0:16]
	case l < 24:
		return bytes.Repeat(b, 3)[0:24]
	case l < 32:
		return bytes.Repeat(b, 2)[0:32]
	default:
		return b[0:32]
	}

}

func Encrypt(key, plaintext []byte) ([]byte, error) {
	key = cutKey(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, err
	}

	ciphered := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphered, nil
}

func Decrypt(key, ciphered []byte) ([]byte, error) {
	key = cutKey(key)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphered) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphered := ciphered[:nonceSize], ciphered[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphered, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}
