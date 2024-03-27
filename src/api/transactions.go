package database

import (
	custom "github.com/Zaikoa/rapid/src/handling"
	"github.com/gogo/protobuf/test/custom"
)

type Transaction struct {
	From_user string
	File_name string
}

/*
retrieves pending file transfer requests for a certain user
*/
func GetPendingTransfers(id int) ([]Transaction, error) {
	query := `
	SELECT nickname, filename
	FROM transfer 
	INNER JOIN users ON transfer.from_user = users.id
	WHERE to_user=$1`
	if rows, err := conn.Query(query, id); err != nil {
		return nil, err
	}
	defer rows.Close()

	// Formats all the transaction details and appends to struct list
	var inbox []Transaction
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.From_user, &transaction.File_name); err != nil {
			return nil, err
		}
		inbox = append(inbox, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return inbox, nil
}

/*
Checks if a user has the right to view a transaction
*/
func UserCanViewTransaction(id int, filename string) bool {
	query := `SELECT id FROM transfer WHERE (from_user = $1 OR to_user = $2) AND filename = $3`
	row := conn.QueryRow(query, id, id, filename)

	var temp int // Temp cannot be null since sql does not support null primary keys
	if err := row.Scan(&temp); err != nil || temp == 0 {
		return false
	}

	return true
}

/*
Deletes a transaction record based on the given id
*/
func DeleteTransaction(id int, filename string) error {
	query := `DELETE FROM transfer WHERE to_user=$1 AND filename=$2`
	if _, err := conn.Exec(query, id, filename); err != nil {
		return err
	}
	return nil
}

/*
Enters the information of a transaction to the database for later referal
*/
func PerformTransaction(from_user_id int, user_to string, filename string, rsd int) error {
	if to_user_id, err := GetUserID(user_to); err != nil {
		return err
	}

	if AreMutualFriends(from_user_id, to_user_id) {
		query := `INSERT INTO transfer (from_user, to_user, filename, rsa_id) VALUES ($1,$2,$3,$4)`
		if _, err := conn.Exec(query, from_user_id, to_user_id, filename, rsd); err != nil {
			return err
		}
		return nil
	}
	return custom.ALREADYFRIENDS
}

/*
Retrieves the key from the database that is used to encrypt the file
*/
func RetrievePublicKey(id int) (string, error) {
	var key string
	query := `SELECT key FROM publickey WHERE users_id = $1`
	if err := conn.QueryRow(query, id).Scan(&key); err != nil {
		return "", err
	}

	return key, nil
}

/*
Inserts the public key into the userkey table
*/
func InsertPublicKey(id int, key string) error {
	query := `INSERT INTO publickey (users_id, key) VALUES ($1, $2)`
	if _, err := conn.Exec(query, id, key); err != nil {
		return err
	}

	return nil
}

/*
Inserts the encrypted nounce and key into rsa table
*/
func InsertRSA(nounce []byte, key []byte) (int, error) {
	var id int
	query := `INSERT INTO rsa (nounce, key) VALUES ($1, $2) RETURNING id`
	if err := conn.QueryRow(query, nounce, key).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

/*
Returns the RSA encrypted nounce and key
*/
func RetrieveRSA(user int, file string) ([]byte, []byte, error) {
	var nounce, key []byte
	query := `
	SELECT nounce, key 
	FROM rsa 
	INNER JOIN transfer ON transfer.rsa_id=rsa.rsa_id

	WHERE to_user=$1 AND filename=$2`
	if err := conn.QueryRow(query, user, file).Scan(&nounce, &key); err != nil {
		return nil, nil, err
	}

	return nounce, key, nil
}
