package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
)

func main() {
	key_pair, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		panic(err)
	}

	// encrypt as pem
	new_private_key := x509.MarshalPKCS1PrivateKey(key_pair)
	new_public_key := x509.MarshalPKCS1PublicKey(&key_pair.PublicKey)
	if err != nil {
		panic(err)
	}

	private_key_pem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: new_private_key,
	})
	public_key_pem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: new_public_key,
	})

	println(string(private_key_pem))
	println(string(public_key_pem))
}
