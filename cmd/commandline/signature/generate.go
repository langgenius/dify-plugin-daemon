package signature

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
)

func GenerateKeyPair(keyPairName string) error {
	// generate a key pair
	keyPair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error("failed to generate key pair", "error", err)
		return err
	}

	// marshal the keys to PEM format
	privateKey := x509.MarshalPKCS1PrivateKey(keyPair)
	publicKey := x509.MarshalPKCS1PublicKey(&keyPair.PublicKey)
	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKey,
	})
	publicKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKey,
	})

	// write the keys to files
	privateKeyPath := fmt.Sprintf("%s.private.pem", keyPairName)
	publicKeyPath := fmt.Sprintf("%s.public.pem", keyPairName)

	if err := os.WriteFile(privateKeyPath, privateKeyPem, 0644); err != nil {
		log.Error("failed to write private key", "error", err)
		return err
	}

	if err := os.WriteFile(publicKeyPath, publicKeyPem, 0644); err != nil {
		log.Error("failed to write public key", "error", err)
		return err
	}

	log.Info("key pair generated successfully", "private_key", privateKeyPath, "public_key", publicKeyPath)

	return nil
}
