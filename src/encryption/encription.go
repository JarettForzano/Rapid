package encription

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	rand2 "crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) []byte {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand2.Reader, pub, msg, nil)
	if err != nil {
		log.Fatal(err)
	}
	return ciphertext
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) []byte {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand2.Reader, priv, ciphertext, nil)
	if err != nil {
		log.Fatal(err)
	}
	return plaintext
}

// Compresses directory or folder into .tar.xz
func Compress(path string, name string) error {
	current_dir, _ := os.Getwd()

	cmd := exec.Command("tar", "-cJf", name, path)
	cmd.Dir = current_dir

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Runs cmd command
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// Decompresses directory or folder and returns it to original state
func Decompress(path string) error {
	current_dir, _ := os.Getwd()

	temp := strings.Split(path, "\\")
	name := temp[len(temp)-1] // Extracts the name of the file

	cmd := exec.Command("tar", "-xf", name)
	cmd.Dir = current_dir

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Runs cmd command
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	}
	return nil
}

// Creates the private key text file for the user
func CreatePrivateEncryptFile(privateKey *rsa.PrivateKey) error {
	err := os.WriteFile("supersecretekey.txt", PrivateKeyToBytes(privateKey), 0644)
	if err != nil {
		return err
	}
	return nil
}

/*
Encrypts location given using public key as a string
*/
func RSAEncryptItem(key string, publickey string, nonce []byte) ([]byte, []byte) {

	public_key_bytes := []byte(publickey) // Reverts key to byte to encrypt with
	publicKey := BytesToPublicKey(public_key_bytes)

	encryptedaes := EncryptWithPublicKey([]byte(key), publicKey)
	encryptedNonce := EncryptWithPublicKey(nonce, publicKey)

	return encryptedaes, encryptedNonce
}

/*
Decrypts file at location given using private key path
*/
func RSADecryptItem(keypath string, aes []byte, nonce []byte) ([]byte, []byte) {
	private_key_bytes, _ := os.ReadFile(keypath)
	privateKey := BytesToPrivateKey(private_key_bytes)

	decryptedAes := DecryptWithPrivateKey(aes, privateKey)

	decryptedNounce := DecryptWithPrivateKey(nonce, privateKey)

	return decryptedAes, decryptedNounce
}

/*
Ecnrypts file at location given using private key path
*/
func AESEncryptionItem(location string, rename string, keyString string) ([]byte, error) {
	key, _ := hex.DecodeString(keyString)

	v, _ := os.ReadFile(location)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	ciphertext := aesgcm.Seal(nil, nonce, v, nil)
	os.WriteFile(rename, ciphertext, 0644)
	return nonce, nil
}

/*
Ecnrypts file at location given using private key path
*/
func AESDecryptItem(location string, rename string, keyString []byte, nonce []byte) error {
	key, _ := hex.DecodeString(string(keyString))

	v, _ := os.ReadFile(location)

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	file, _ := aesgcm.Open(nil, nonce, v, nil)
	os.WriteFile(rename, file, 0644)
	return nil
}
