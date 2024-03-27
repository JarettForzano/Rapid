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
	"os"
	"os/exec"
	"path/filepath"

	custom "github.com/Zaikoa/rapid/src/handling"
)

// EncryptWithPublicKey encrypts data with public key
func EncryptWithPublicKey(msg []byte, pub *rsa.PublicKey) ([]byte, error) {
	hash := sha512.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand2.Reader, pub, msg, nil)
	if err != nil {
		return nil, custom.NewError(err.Error())
	}
	return ciphertext, nil
}

// DecryptWithPrivateKey decrypts data with private key
func DecryptWithPrivateKey(ciphertext []byte, priv *rsa.PrivateKey) ([]byte, error) {
	hash := sha512.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand2.Reader, priv, ciphertext, nil)
	if err != nil {
		return nil, custom.NewError(err.Error())
	}
	return plaintext, nil
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

	if err := cmd.Run(); err != nil {
		return custom.NewError(fmt.Sprint(err) + ": " + stderr.String())
	}

	return nil
}

// Decompresses directory or folder and returns it to original state
func Decompress(path string) error {
	current_dir, _ := os.Getwd()

	name := filepath.Base(path) // Extracts the name of the file

	cmd := exec.Command("tar", "-xf", name)
	cmd.Dir = current_dir

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Runs cmd command
	if err := cmd.Run(); err != nil {
		return custom.NewError(fmt.Sprint(err) + ": " + stderr.String())
	}

	return nil
}

// Creates the private key text file for the user
func CreatePrivateEncryptFile(privateKey *rsa.PrivateKey) error {
	if err := os.WriteFile("supersecretekey.txt", PrivateKeyToBytes(privateKey), 0644); err != nil {
		return err
	}

	return nil
}

/*
Encrypts location given using public key as a string
*/
func RSAEncryptItem(key string, publickey string, nonce []byte) ([]byte, []byte, error) {

	public_key_bytes := []byte(publickey) // Reverts key to byte to encrypt with
	publicKey := BytesToPublicKey(public_key_bytes)

	encryptedaes, err := EncryptWithPublicKey([]byte(key), publicKey)
	if err != nil {
		return nil, nil, err
	}
	encryptedNonce, err := EncryptWithPublicKey(nonce, publicKey)
	if err != nil {
		return nil, nil, err
	}
	return encryptedaes, encryptedNonce, nil
}

/*
Decrypts file at location given using private key path
*/
func RSADecryptItem(keypath string, aes []byte, nonce []byte) ([]byte, []byte, error) {
	private_key_bytes, _ := os.ReadFile(keypath)
	privateKey := BytesToPrivateKey(private_key_bytes)

	decryptedAes, err := DecryptWithPrivateKey(aes, privateKey)
	if err != nil {
		return nil, nil, err
	}
	decryptedNounce, err := DecryptWithPrivateKey(nonce, privateKey)
	if err != nil {
		return nil, nil, err
	}
	return decryptedAes, decryptedNounce, nil
}

/*
Ecnrypts file at location given using private key path
*/
func AESEncryptionItem(location string, rename string, keyString string) ([]byte, error) {
	key, _ := hex.DecodeString(keyString)

	bytes, err := os.ReadFile(location)
	if err != nil {
		return nil, err
	}
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

	ciphertext := aesgcm.Seal(nil, nonce, bytes, nil)
	os.WriteFile(rename, ciphertext, 0644)
	return nonce, nil
}

/*
Ecnrypts file at location given using private key path
*/
func AESDecryptItem(location string, rename string, keyString []byte, nonce []byte) error {
	key, _ := hex.DecodeString(string(keyString))

	bytes, err := os.ReadFile(location)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	file, _ := aesgcm.Open(nil, nonce, bytes, nil)
	err = os.WriteFile(rename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}
