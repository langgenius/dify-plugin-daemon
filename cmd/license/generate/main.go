package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	keyPair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	// encrypt as pem
	newPrivateKey := x509.MarshalPKCS1PrivateKey(keyPair)
	newPublicKey := x509.MarshalPKCS1PublicKey(&keyPair.PublicKey)
	if err != nil {
		panic(err)
	}

	privateKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: newPrivateKey,
	})
	publicKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: newPublicKey,
	})

	os.WriteFile("internal/core/license/private_key/PRIVATE_KEY.pem", privateKeyPem, 0644)
	os.WriteFile("internal/core/license/public_key/PUBLIC_KEY.pem", publicKeyPem, 0644)
}
