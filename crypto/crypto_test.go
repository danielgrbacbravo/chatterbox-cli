package crypto_test

import (
	"chatterbox-cli/crypto"
	"testing"

	log "github.com/charmbracelet/log"
)

// TestGeneratePrivateKey tests the GeneratePrivateKey function
func TestGeneratePrivateKey(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	if privateKey == nil {
		t.Error("GeneratePrivateKey() returned nil")
	}
}

// TestGeneratePublicKey tests the GeneratePublicKey function
func TestGeneratePublicKey(t *testing.T) {
	privateKey := crypto.GeneratePrivateKey()
	publicKey := crypto.GeneratePublicKey(privateKey)
	if publicKey.X == nil || publicKey.Y == nil {
		t.Error("GeneratePublicKey() returned nil")
	}
}

// TestGenerateSharedSecret tests the GenerateSharedSecret function
func TestGenerateSharedSecret(t *testing.T) {
	privateKey1 := crypto.GeneratePrivateKey()
	privateKey2 := crypto.GeneratePrivateKey()
	publicKey1 := crypto.GeneratePublicKey(privateKey1)
	publicKey2 := crypto.GeneratePublicKey(privateKey2)
	sharedSecret1 := crypto.GenerateSharedSecret(privateKey1, publicKey2)
	sharedSecret2 := crypto.GenerateSharedSecret(privateKey2, publicKey1)
	if sharedSecret1.Cmp(sharedSecret2) != 0 {
		t.Error("GenerateSharedSecret() returned different shared secrets")
	}
}

// TestEncryptDecrypt tests the Encrypt and Decrypt functions
func TestEncryptDecrypt(t *testing.T) {
	privateKey1 := crypto.GeneratePrivateKey()
	privateKey2 := crypto.GeneratePrivateKey()
	publicKey1 := crypto.GeneratePublicKey(privateKey1)
	publicKey2 := crypto.GeneratePublicKey(privateKey2)
	sharedSecret1 := crypto.GenerateSharedSecret(privateKey1, publicKey2)
	sharedSecret2 := crypto.GenerateSharedSecret(privateKey2, publicKey1)
	// check if shared secrets are the same
	if sharedSecret1.Cmp(sharedSecret2) != 0 {
		t.Error("GenerateSharedSecret() returned different shared secrets")
	} else {
		log.Info("Matched shared secrets")
	}

	msg := "Hello, world!"
	log.Info("Original Message", "msg", msg)
	encryptedMsg := crypto.Encrypt(msg, sharedSecret1)
	log.Info("encryptedMsg: ", "encrMsg", encryptedMsg)
	decryptedMsg := crypto.Decrypt(encryptedMsg, sharedSecret1)
	log.Info("decryptedMsg: ", "decrMsg", decryptedMsg)
	if decryptedMsg != msg {
		t.Error("Encrypt()/Decrypt() returned different messages")
	}
}
