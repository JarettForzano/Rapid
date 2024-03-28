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

	byte_result, err := encription.Compress(filesource)
	if err != nil {
		return err
	}

	nonce, err := encription.AESEncryptionItem(compressed_name, byte_result, AESkey)
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

	if err = cloud.UploadToMega(compressed_name, user, to_user); err != nil {
		return err
	}

	// Deletes encypted file in temp
	if err = os.RemoveAll(filepath.Join(os.TempDir(), compressed_name)); err != nil {
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
	if !database.UserCanViewTransaction(user, file) {
		return custom.TRANSACTIONNOTEXIST
	}

	user_name, err := database.GetUserNameByID(user)
	if err != nil {
		return err
	}
	encodedname := database.HashInfo(file + user_name)

	compressed_name := fmt.Sprintf("%s.tar.xz", encodedname)
	err = cloud.DownloadFromMega(compressed_name)
	if err != nil { // Returns bytes of encrypted file
		return err
	}

	// Removes the copy from the cloud so that no users can access it
	if err := cloud.DeleteFromMega(user, compressed_name); err != nil {
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

	if err = encription.AESDecryptItem(compressed_name, KeyD, NonceD); err != nil { // Pass in bytes of encrypted file and then the file will be decrypted and put in the temp folder
		return err
	}

	if err = encription.Decompress(file, location, compressed_name); err != nil { // Takes in the name of the decrypted file and then the file is decrypted into location passed in
		return err
	}

	if err = database.DeleteTransaction(user, file); err != nil {
		return err
	}

	// deletes encrypyted file
	if err = os.RemoveAll(filepath.Join(os.TempDir(), compressed_name)); err != nil {
		log.Println(err)
	}

	return nil
}
