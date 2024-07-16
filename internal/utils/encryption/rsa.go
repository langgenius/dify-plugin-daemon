package encryption

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io"
)

func RSASign(rsaPrivateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed[:])
}

func VerifySign(rsaPublicKey *rsa.PublicKey, data []byte, sign []byte) error {
	hashed := sha256.Sum256(data)
	return rsa.VerifyPKCS1v15(rsaPublicKey, crypto.SHA256, hashed[:], sign)
}

func AESEncrypt(aesKey []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	cipherText := aesGCM.Seal(nonce, nonce, data, nil)
	return cipherText, nil
}

func AESDecrypt(aesKey []byte, data []byte) ([]byte, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, cipherText := data[:nonceSize], data[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return nil, err
	}

	return plainText, nil
}

func LoadPrivateKey(data []byte) (*rsa.PrivateKey, error) {
	private_key_block, rest := pem.Decode(data)
	if private_key_block == nil || private_key_block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("failed to decode PEM block containing private key")
	}

	if len(rest) != 0 {
		return nil, errors.New("extra data included in the PEM block")
	}

	private_key, err := x509.ParsePKCS1PrivateKey(private_key_block.Bytes)
	if err != nil {
		return nil, err
	}

	return private_key, nil
}

func LoadPublicKey(data []byte) (*rsa.PublicKey, error) {
	public_key_block, rest := pem.Decode(data)
	if public_key_block == nil || public_key_block.Type != "RSA PUBLIC KEY" {
		return nil, errors.New("failed to decode PEM block containing public key")
	}

	if len(rest) != 0 {
		return nil, errors.New("extra data included in the PEM block")
	}

	public_key, err := x509.ParsePKCS1PublicKey(public_key_block.Bytes)
	if err != nil {
		return nil, err
	}

	return public_key, nil
}
