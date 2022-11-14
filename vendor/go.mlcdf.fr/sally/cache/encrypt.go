package cache

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

func encrypt(text []byte, key []byte) ([]byte, error) {
	if len(text) == 0 {
		return []byte{}, nil
	}

	cphr, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcm, err := cipher.NewGCM(cphr)
	if err != nil {
		return []byte{}, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return []byte{}, err
	}
	return gcm.Seal(nonce, nonce, text, nil), nil
}

func decrypt(text []byte, key []byte) ([]byte, error) {
	if len(text) == 0 {
		return []byte{}, nil
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}
	gcmDecrypt, err := cipher.NewGCM(c)
	if err != nil {
		return []byte{}, err
	}
	nonceSize := gcmDecrypt.NonceSize()
	if len(text) < nonceSize {
		return []byte{}, err
	}
	nonce, encryptedMessage := text[:nonceSize], text[nonceSize:]
	plaintext, err := gcmDecrypt.Open(nil, nonce, encryptedMessage, nil)
	if err != nil {
		return []byte{}, err
	}
	return plaintext, nil
}
