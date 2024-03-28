package error

import (
	"errors"
	"fmt"
	"runtime"
)

func Debugging(message string) error {
	_, file, line, _ := runtime.Caller(1)
	return fmt.Errorf("[%s][%d] : %s", file, line, message)
}

func NewError(message string) error {
	return errors.New(message)
}

var (
	INVALIDKEY          = errors.New("Key is not located at this spot or is not correct")
	ALREADYFRIENDS      = errors.New("You cannot add a player you are already friends with")
	USERNOTEXIST        = errors.New("This user either does not exist or the friend code is incorrect")
	FAILEDCONNECTION    = errors.New("Failed to connected to database - check credentials")
	NOTLOGGEDIN         = errors.New("Cannot use main methods without logging in")
	USERTAKEN           = errors.New("Username hsa been taken")
	NOTFOUND            = errors.New("User not found")
	WRONGINFO           = errors.New("Username or password is wrong")
	TRANSACTIONNOTEXIST = errors.New("Transaction does not exist")
)
