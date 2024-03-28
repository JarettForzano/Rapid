package cloud

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	custom "github.com/Zaikoa/rapid/src/handling"
)

/*
Uploads a zip to the cloud
*/
func UploadToMega(name string, from_user_id int, user_to string) error {
	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	// Sends that file to MEGA
	cmd := exec.Command("megacmd", config, "put", name, "mega:/")
	cmd.Dir = os.TempDir()

	// Runs cmd command
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func DownloadFromMega(file string) error {
	// Formats it for the mega cloud (readjusts the name to fit the hashing)
	cloud_dir := fmt.Sprintf("mega:/%s", file)

	// Handles megacmd config
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.megacmd.json")
	config := fmt.Sprintf(`-conf=%s`, directory)

	location := filepath.Join(os.TempDir(), file)
	cmd := exec.Command("megacmd", config, "get", cloud_dir, location)

	err := cmd.Run()
	if err != nil {
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
