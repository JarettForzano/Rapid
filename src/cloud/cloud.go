package cloud

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	database "github.com/Zaikoa/rapid/src/api"
	custom "github.com/Zaikoa/rapid/src/handling"
)

/*
Uploads a zip to the cloud
*/
func UploadToMega(path string, from_user_id int, user_to string) error {
	// Formats the file
	encrypted_name := filepath.Base(path)

	current_dir, err := os.Getwd()
	if err != nil {
		return err
	}

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Sends that file to MEGA
	cmd := exec.Command("megacmd", config, "put", encrypted_name, "mega:/")

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

func DownloadFromMega(user int, original string, file string, location string) error {
	if !database.UserCanViewTransaction(user, original) {
		return custom.TRANSACTIONNOTEXIST
	}

	// Gets the current directory the user is in
	current_dir, _ := os.Getwd()
	destination := filepath.Join(current_dir, location, file)

	// Formats it for the mega cloud (readjusts the name to fit the hashing)
	cloud_dir := fmt.Sprintf("mega:/%s", file)

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	cmd := exec.Command("megacmd", config, "get", cloud_dir, destination)
	cmd.Dir = current_dir

	// Error handing
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return custom.NewError(fmt.Sprint(err) + ": " + stderr.String())
	}

	// Removes the copy from the cloud so that no users can access it
	if err := DeleteFromMega(user, file); err != nil {
		return err
	}

	return nil
}

// Removes the file from the cloud
func DeleteFromMega(user int, file string) error {
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

	if err := cmd.Run(); err != nil {
		return custom.NewError(fmt.Sprint(err) + ": " + stderr.String())
	}

	return nil

}
