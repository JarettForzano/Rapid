package cloud

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	database "github.com/Zaikoa/rapid/src/api"
	custom "github.com/Zaikoa/rapid/src/handling"
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

	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Sends that file to MEGA
	cmd := exec.Command("megacmd", config, "put", path, "mega:/")

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

	return nil
}

func DownloadFromMega(user int, file string, location string) error {

	if user == 0 {
		return custom.NewError("User must be logged in to use this method")
	}

	if !database.UserCanViewTransaction(user, file) {
		return nil
	}

	// Gets the current directory the user is in
	current_dir, _ := os.Getwd()
	stored_name := fmt.Sprintf("%s.tar.xz", HashInfo(file+database.GetUserNameByID(user)))
	destination := filepath.Join(current_dir, location, stored_name)

	// Formats it for the mega cloud (readjusts the name to fit the hashing)
	cloud_dir := fmt.Sprintf("mega:/%s", stored_name)

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

	// Removes the copy from the cloud so that no users can access it
	_, err = DeleteFromMega(user, stored_name)
	if err != nil {
		return err
	}
	return nil
}

// Removes the file from the cloud
func DeleteFromMega(user int, file string) (bool, error) {

	if user == 0 {
		return false, custom.NewError("User must be logged in to use this method")
	}

	// Formats it for the mega cloud
	cloud_dir := fmt.Sprintf("mega:/%s", file)

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

	return true, nil

}
