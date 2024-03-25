package transaction

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	database "github.com/Zaikoa/rapid/src/api"
	"github.com/Zaikoa/rapid/src/cloud"
	encription "github.com/Zaikoa/rapid/src/encryption"
)

// Generates a random 32 character string for encryption purposes
func GenerateKey() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
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
	temp := strings.Split(filesource, "\\")
	name := temp[len(temp)-1] // Extracts the name of the file

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

	publickey, err := database.RetrievePublicKey(user)
	if err != nil {
		return err
	}

	KeyE, NonceE := encription.RSAEncryptItem(AESkey, publickey, nonce)
	id, err := database.InsertRSA(NonceE, KeyE)
	if err != nil {
		return err
	}

	err = database.PerformTransaction(user, to_user, name, id)
	if err != nil {
		return err
	}

	err = cloud.UploadToMega(encrypted_location, user, to_user)
	if err != nil {
		return err
	}

	// Deletes zip folder
	err = os.RemoveAll(encrypted_location)
	if err != nil {
		log.Println(err)
	}
	return nil
}

/*
Recieves a file, decrypts it, and then uncompresses it
*/
func RecieveDecrypt(user int, keypath string, file string, location string) error {

	encodedname := database.HashInfo(file + database.GetUserNameByID(user))
	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}
	compressed_name := fmt.Sprintf("%s.tar.xz", encodedname)
	err = cloud.DownloadFromMega(user, file, compressed_name, location)
	if err != nil {
		return err
	}

	NonceE, KeyE, err := database.RetrieveRSA(user, file)
	if err != nil {
		return err
	}
	KeyD, NonceD := encription.RSADecryptItem(keypath, KeyE, NonceE)

	// Works up until here
	decrypt_here := filepath.Join(current_dir, compressed_name)
	err = encription.AESDecryptItem(decrypt_here, file, KeyD, NonceD)
	if err != nil {
		return err
	}
	location = filepath.Join(current_dir, file)
	err = encription.Decompress(location)
	if err != nil {
		return err
	}
	//err = database.DeleteRSA(user, file)
	if err != nil {
		return err
	}
	//err = database.DeleteTransaction(user, file)
	if err != nil {
		return err
	}
	return nil
}
