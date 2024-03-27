package transaction

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	database "github.com/Zaikoa/rapid/src/api"
	"github.com/Zaikoa/rapid/src/cloud"
	encription "github.com/Zaikoa/rapid/src/encryption"
	custom "github.com/Zaikoa/rapid/src/handling"
)

// Generates a random 32 character string for encryption purposes
func GenerateKey() (string, error) {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes), nil
}

/*
Compresses a file and encrypts it and then sends it to the cloud
*/
func EncryptSend(filesource string, user int, to_user string) error {
	AESkey, err := GenerateKey()
	if err != nil {
		return err
	}
	name := filepath.Base(filesource) // Extracts the name of the file

	encodedname := database.HashInfo(name + to_user)
	compressed_name := fmt.Sprintf("%s.tar.xz", encodedname)

	encription.Compress(filesource, compressed_name)

	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}
	encrypted_location := filepath.Join(current_dir, compressed_name)

	nonce, err := encription.AESEncryptionItem(encrypted_location, compressed_name, AESkey)
	if err != nil {
		return err
	}
	to_user_id, err := database.GetUserID(to_user)
	if err != nil {
		return err
	}
	publickey, err := database.RetrievePublicKey(to_user_id)
	if err != nil {
		return err
	}

	KeyE, NonceE, err := encription.RSAEncryptItem(AESkey, publickey, nonce)
	if err != nil {
		return err
	}

	rsa_id, err := database.PerformTransaction(user, to_user, name)
	if err != nil {
		return err
	}

	if err := database.InsertRSA(NonceE, KeyE, rsa_id); err != nil {
		return err
	}

	if err = cloud.UploadToMega(encrypted_location, user, to_user); err != nil {
		return err
	}

	// Deletes compressed file
	if err = os.RemoveAll(encrypted_location); err != nil {
		log.Println(err)
	}

	return nil
}

/*
Recieves a file, decrypts it, and then uncompresses it
*/
func RecieveDecrypt(user int, keypath string, file string, location string) error {
	if _, err := os.Stat(keypath); err != nil {
		return custom.INVALIDKEY
	}
	user_name, err := database.GetUserNameByID(user)
	if err != nil {
		return err
	}
	encodedname := database.HashInfo(file + user_name)
	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}
	compressed_name := fmt.Sprintf("%s.tar.xz", encodedname)
	if err = cloud.DownloadFromMega(user, file, compressed_name, location); err != nil {
		return err
	}

	NonceE, KeyE, err := database.RetrieveRSA(user, file)

	if err != nil {
		return err
	}
	KeyD, NonceD, err := encription.RSADecryptItem(keypath, KeyE, NonceE)
	if err != nil {
		return err
	}

	decrypt_here := filepath.Join(current_dir, compressed_name)

	if err = encription.AESDecryptItem(decrypt_here, compressed_name, KeyD, NonceD); err != nil {
		return err
	}

	if err = encription.Decompress(decrypt_here); err != nil {
		return err
	}

	if err = database.DeleteTransaction(user, file); err != nil {
		return err
	}

	// deletes encrypyted file
	if err = os.RemoveAll(decrypt_here); err != nil {
		return custom.NewError(err.Error())
	}

	return nil
}
