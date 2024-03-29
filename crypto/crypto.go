package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"go-chat-cli/message"
	"io"
	"log"
	"math/big"
	"net"
	"strings"
)

// using the elliptic curve digital signature algorithm (ECDSA) to generate a private key
// and the diffie-hellman key exchange algorithm to generate a shared secret

// / GeneratePrivateKey generates a private key
func GeneratePrivateKey() *ecdsa.PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatal(err)
	}
	return privateKey
}

// / GeneratePublicKey generates a public key from a private key
func GeneratePublicKey(privateKey *ecdsa.PrivateKey) ecdsa.PublicKey {
	return privateKey.PublicKey
}

// / GenerateSharedSecret generates a shared secret from a private key and a public key
func GenerateSharedSecret(privateKey *ecdsa.PrivateKey, publicKey ecdsa.PublicKey) *big.Int {
	x, _ := privateKey.PublicKey.Curve.ScalarMult(publicKey.X, publicKey.Y, privateKey.D.Bytes())
	return x
}

func generateKey(secret *big.Int) []byte {
	// Convert shared secret to byte slice
	secretBytes := secret.Bytes()

	// Create a new hash.
	h := sha256.New()

	// Write secretBytes to the hash.
	h.Write(secretBytes)

	// Sum up the hash to the key.
	key := h.Sum(nil)

	return key
}
func Decrypt(securedMessage string, secret *big.Int) (decodedmess string) {
	ciphertext, _ := base64.URLEncoding.DecodeString(securedMessage)

	key := generateKey(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(ciphertext) < aes.BlockSize {
		panic("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	decodedmess = string(ciphertext)
	return
}

func SendPublicKey(conn net.Conn, publicKey ecdsa.PublicKey) {
	message := message.Message{
		MessageType: "public_key",
		Message:     publicKey.X.String() + "," + publicKey.Y.String(),
	}
	message.SendMessage(conn)
}

func ConvertToPublicKey(message message.Message) ecdsa.PublicKey {
	xySlice := strings.Split(message.Message, ",")
	x, _ := new(big.Int).SetString(xySlice[0], 10)
	y, _ := new(big.Int).SetString(xySlice[1], 10)

	publicKey := ecdsa.PublicKey{
		X: x,
		Y: y,
	}
	return publicKey
}

func ReceivePublicKey(conn net.Conn) ecdsa.PublicKey {
	msg, err := message.ReadMessage(conn)
	if err != nil {
		log.Fatal(err)
	}
	xy := msg.Message
	xySlice := strings.Split(xy, ",")
	x, _ := new(big.Int).SetString(xySlice[0], 10)
	y, _ := new(big.Int).SetString(xySlice[1], 10)

	publicKey := ecdsa.PublicKey{
		X: x,
		Y: y,
	}
	return publicKey
}

func Encrypt(message string, secret *big.Int) (encmess string) {
	plaintext := []byte(message)

	key := generateKey(secret)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	// convert to base64
	encmess = base64.URLEncoding.EncodeToString(ciphertext)
	return
}

func EncyptMessage(message message.Message, secret *big.Int) message.Message {
	message.Username = Encrypt(message.Username, secret)
	message.Message = Encrypt(message.Message, secret)
	return message
}

func DecryptMessage(message message.Message, secret *big.Int) message.Message {
	message.Username = Decrypt(message.Username, secret)
	message.Message = Decrypt(message.Message, secret)
	return message
}

func BroadcastMessage(message message.Message, clients []*net.Conn, sharedKeys map[net.Conn]*big.Int) {
	for _, client := range clients {
		message := EncyptMessage(message, sharedKeys[*client])
		message.SendMessage(*client)
	}
}
