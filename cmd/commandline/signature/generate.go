package signature

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
)

func GenerateKeyPair(keyPairName string) error {
	// generate a key pair
	keyPair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Error("Failed to generate key pair: %v", err)
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
		log.Error("Failed to write private key: %v", err)
		return err
	}

	if err := os.WriteFile(publicKeyPath, publicKeyPem, 0644); err != nil {
		log.Error("Failed to write public key: %v", err)
		return err
	}

	log.Info("Key pair generated successfully")
	log.Info("Private key: %s", privateKeyPath)
	log.Info("Public key: %s", publicKeyPath)

	return nil
}
