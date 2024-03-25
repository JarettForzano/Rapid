package cloud

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	database "github.com/Zaikoa/rapid/src/api"
	custom "github.com/Zaikoa/rapid/src/handling"
	encription "github.com/Zaikoa/rapid/src/rsa"
)

// Generates a random 32 character string for encryption purposes
func GenerateKey() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

/*
Uploads a zip to the cloud
*/
func UploadToMega(path string, from_user_id int, user_to string) error {

	if from_user_id == 0 {
		return custom.NewError("User must be logged in to use this method")
	}
	// Formats the file
	split := strings.Split(path, "/")
	name_of_item := split[len(split)-1]

	// Makes sure user is allowed to send the file before procceding
	result, err := database.PerformTransaction(from_user_id, user_to, name_of_item)
	if err != nil {
		return err
	}
	if !result {
		return nil
	}

	// makes name include the file
	name := database.HashInfo(name_of_item + user_to)
	err = encription.Compress(path, name)
	if err != nil {
		return err
	}

	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}

	name = fmt.Sprintf("%s.tar.xz", name)
	send_me := filepath.Join(current_dir, name)
	id, err := database.GetUserID(user_to)
	if err != nil {
		return err
	}
	key, err := database.RetrieveKey(id)
	if err != nil {
		return err
	}
	err = encription.EncryptItem(send_me, key)

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Sends that file to MEGA
	cmd := exec.Command("megacmd", config, "put", send_me, "mega:/")

	cmd.Dir = current_dir

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Runs cmd command
	err = cmd.Run()
	if err != nil {
		return err
	}

	// Deletes zip folder
	err = os.RemoveAll(send_me)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func DownloadFromMega(user int, file_name string, location string, privatekeypath string) error {

	if user == 0 {
		return custom.NewError("User must be logged in to use this method")
	}

	if !database.UserCanViewTransaction(user, file_name) {
		return nil
	}

	// Gets the current directory the user is in
	current_dir, _ := os.Getwd()
	name := database.HashInfo(file_name + database.GetUserNameByID(user))
	encryped_name := fmt.Sprintf("%s.tar.xz", name)
	filename := fmt.Sprintf("%s.tar.xz", file_name)
	// Destination the file will be downloaded to
	destination := filepath.Join(current_dir, location, filename)

	// Formats it for the mega cloud (readjusts the name to fit the hashing)
	cloud_dir := fmt.Sprintf("mega:/%s", encryped_name)

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Calls cmd command to retrieve the file
	cmd := exec.Command("megacmd", config, "get", cloud_dir, destination)
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
	}

	// Decripts folder
	err = encription.DecryptItem(destination, privatekeypath)
	if err != nil {
		return err
	}
	err = encription.Decompress(destination)
	if err != nil {
		return err
	}
	// Deletes zip folder
	err = os.RemoveAll(destination)
	if err != nil {
		return err
	}

	// Removes the copy from the cloud so that no users can access it
	_, err = DeleteFromMega(user, file_name)
	if err != nil {
		return err
	}
	return nil
}

// Removes the file from the cloud
func DeleteFromMega(user int, file_name string) (bool, error) {

	if user == 0 {
		return false, custom.NewError("User must be logged in to use this method")
	}

	name := database.HashInfo(file_name + database.GetUserNameByID(user))

	// Formats it for the mega cloud
	cloud_dir := fmt.Sprintf("mega:/%s.tar.xz", name)

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Calls cmd command to retrieve the file
	cmd := exec.Command("megacmd", config, "delete", cloud_dir)

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	// Runs cmd command
	err := cmd.Run()
	if err != nil {
		return false, err
	}
	err = database.DeleteTransaction(user, file_name)

	if err != nil {
		return false, nil
	}
	return true, nil

}
