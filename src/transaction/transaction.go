package transaction

import (
	"crypto/rand"
	"encoding/hex"
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
func EncryptSend(filesource string, id int, to_user string) error {
	AESkey := GenerateKey()

	temp := strings.Split(path, "/")
	name := temp[len(temp)-1] // Extracts the name of the file

	encodedname := HashInfo(name + database.GetUserNameByID(id))
	encription.Compress(filesource, encodedname)

	current_dir := os.Getwd()
	encrypted_location := filepath.Join(current_dir, encodedname)

	nonce, err := encription.AESEncryptionItem(encrypted_location, encodedname)
	if err != nil {
		return err
	}

	publickey := database.RetrievePublicKey(id)

	KeyE, NonceE := encription.RSAEncryptItem(AESkey, publickey, nonce)

	id, err = database.InsertRSA(NonceE, KeyE)

	err = database.PerformTransaction(id, to_user, name)
	if err != nil {
		return err
	}

	err = cloud.UploadToMega(encrypted_location, id, database.GetUserNameByID(to_user))
	if err != nil {
		return err
	}
	return nil
}

/*
Recieves a file, decrypts it, and then uncompresses it
*/
func RecieveDecrypt(user int, keypath string, file string, location string) error {
	err := cloud.DownloadFromMega(user, file, location)
	if err != nil {
		return err
	}

	NonceE, KeyE := datbase.RetrieveRSA(user, file)
	KeyD, NonceD := encription.RSADecryptItem(keypath, KeyE, NonceE)

	temp := strings.Split(path, "/")
	name := temp[len(temp)-1] // Extracts the name of the file

	encodedname := HashInfo(name + database.GetUserNameByID(id))
	current_dir := os.Getwd()

	decrypt_here := filepath.Join(current_dir, encodedname)
	err = encription.AESDecryptItem(decrypt_here, file, KeyD, NonceD)
	if err != nil {
		return err
	}
	location := filepath.Join(current_dir, file)
	err := encription.Decompress(location)
	if err != nil {
		return err
	}
	err = database.DeleteRSA(user, file)
	if err != nil {
		return err
	}
	err = database.DeleteTransaction(user, file)
	if err != nil {
		return err
	}
	return nil
}
