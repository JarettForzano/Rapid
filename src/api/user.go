package database

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	encription "github.com/Zaikoa/rapid/src/encryption"
	custom "github.com/Zaikoa/rapid/src/handling"
	"github.com/jackc/pgx"
)

// Stores the current users id
var current_user int

/*
Returns users UUID by first checking what system they are on
*/
func getUUID() (string, error) {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("wmic", "path", "win32_computersystemproduct", "get", "UUID")
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		out := string(b)
		result := strings.Split(out, "\n")

		return result[1], nil
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("uuidgen")
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		out := string(b)

		return out, nil
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("findmnt", "/", "-o", "UUID", "-n")
		b, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		out := string(b)

		return out, nil
	} else {
		return "", nil
	}
}

/*
Retrieves a user's name based on their id, which is passed in
*/
func GetUserNameByID(id int) (string, error) {
	var name string
	query := `SELECT username FROM users WHERE id=$1`
	if err := conn.QueryRow(query, id).Scan(&name); err != nil {
		return "", custom.NOTFOUND
	}

	return name, nil
}

/*
Generates a random friend code that will be assigned to a user during account creation
*/
func generateFriendCode() string {
	rand.Seed(time.Now().UnixNano())
	allowedChars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// Generate a random 8-character string
	var result string
	for i := 0; i < 8; i++ {
		result += string(allowedChars[rand.Intn(len(allowedChars))])
	}

	return result
}

/*
Hashes using the SHA256 package
*/
func HashInfo(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	z := h.Sum(nil)
	hashString := hex.EncodeToString(z)

	return hashString
}

/*
Creates an account in the database
uuid is unique to the computer, and is used on startup to indentify the device
*/
func CreateAccount(username string, password string) error {
	password = HashInfo(password)
	uuid, err := getUUID()
	if err != nil {
		return err
	}

	uuid = HashInfo(uuid)
	result, _ := GetUserID(username)
	if result != 0 {
		return custom.USERTAKEN
	}
	code := generateFriendCode()

	fmt.Print("Enter in nickname >> ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	text := scanner.Text()
	nickname := strings.Trim(text, "\n") // Trims newline from extracted nickname

	// Inserts that data inside of the datbase
	query := `INSERT INTO users (username, nickname, password, friend_code, uuid) VALUES ($1, $2, $3, $4, $5)`
	if _, err := conn.Exec(query, username, nickname, password, code, uuid); err != nil {
		return err
	}

	privateKey, publicKey := encription.GenerateKeyPair(4098)
	if err := encription.CreatePrivateEncryptFile(privateKey); err != nil {
		return err
	}

	id, err := GetUserID(username)
	if err != nil {
		return err
	}

	if err = InsertPublicKey(id, string(encription.PublicKeyToBytes(publicKey))); err != nil {
		return err
	}
	fmt.Println("Congrats! Heres your decription key. Dont lose it.....")

	return nil
}

/*
Retrieves a user's freind code based on their name, which is passed in
*/
func GetUserFriendCode(id int) (string, error) {
	var code string
	query := `SELECT friend_code FROM users WHERE id=$1`
	if err := conn.QueryRow(query, id).Scan(&code); err != nil {
		return "", err
	}

	return code, nil
}

/*
Retrieves a user's id based on their name, which is passed in
*/
func GetUserID(name string) (int, error) {
	var id int
	query := `SELECT id FROM users WHERE username=$1`
	if err := conn.QueryRow(query, name).Scan(&id); err != nil || id == 0 {
		return id, custom.NOTFOUND
	}

	return id, nil
}

/*
Sets the user who is loggin in
*/
func Login(username string, password string) error {
	uuid, err := getUUID()
	if err != nil {
		return err
	}
	uuid = HashInfo(uuid)

	password = HashInfo(password)
	query := `SELECT id FROM users WHERE username=$1 AND password=$2`
	if err = conn.QueryRow(query, username, password).Scan(&current_user); err != nil || current_user == 0 {
		return custom.WRONGINFO
	}
	if err = deactivateSessions(uuid); err != nil {
		return err
	}

	query = `UPDATE users SET session = 1, uuid=$1 WHERE username=$2` // Sets the session to active
	if _, err = conn.Exec(query, uuid, username); err != nil {
		return err
	}

	return nil
}

// Returns the current users id and sets the session
func SetActiveSession() (int, error) {
	uuid, err := getUUID()
	if err != nil {
		return current_user, err
	}

	uuid = HashInfo(uuid)
	query := `SELECT id FROM users WHERE uuid=$1 AND session=1`
	if err := conn.QueryRow(query, uuid).Scan(&current_user); err != nil {
		if err == pgx.ErrNoRows {
			return current_user, nil
		}
		return current_user, err
	}

	return current_user, nil
}

func GetCurrentId() int {
	return current_user
}

/*
Deactivates the session of a single user
*/
func DeactivateSession(id int) error {
	query := `UPDATE users SET session=0 WHERE id=$1`
	if _, err := conn.Exec(query, id); err != nil {
		return err
	}

	return nil
}

/*
Deactivates all other user session on teh same device
*/
func deactivateSessions(uuid string) error {
	query := `UPDATE users SET session=0 WHERE uuid=$1`
	if _, err := conn.Exec(query, uuid); err != nil {
		return err
	}

	return nil
}
