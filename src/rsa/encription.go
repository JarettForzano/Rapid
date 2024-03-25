package encription

import (
	"bytes"
	rand2 "crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	compressed_name := fmt.Sprintf("%s.tar.xz", name)

	cmd := exec.Command("tar", "-cJf", compressed_name, path)
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
	file_location := filepath.Join(current_dir, path)

	cmd := exec.Command("tar", "-xf", file_location)
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
func EncryptItem(path string, publickey string) error {
	temp := strings.Split(path, "/")
	filename := temp[len(temp)-1] // Extracts the name of the file

	v, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	public_key_bytes := []byte(publickey) // Reverts key to byte to encrypt with
	publicKey := BytesToPublicKey(public_key_bytes)
	fmt.Println(string(PublicKeyToBytes(publicKey)))
	encryptedText := EncryptWithPublicKey(v, publicKey)
	err = os.WriteFile(filename, encryptedText, 0644)
	if err != nil {
		return err
	}
	return nil
}

/*
Decrypts file at location given using private key path
*/
func DecryptItem(path string, privatekeypath string) error {
	temp := strings.Split(path, "/")
	filename := temp[len(temp)-1] // Extracts the name of the file

	v, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	privatekey, err := os.ReadFile(privatekeypath)
	privateKey := BytesToPrivateKey(privatekey)

	decryptedText := DecryptWithPrivateKey(v, privateKey)
	err = os.WriteFile(filename, decryptedText, 0644)
	if err != nil {
		return err
	}
	return nil
}
