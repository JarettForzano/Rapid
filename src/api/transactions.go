package database

import (
	custom "github.com/Zaikoa/rapid/src/handling"
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
	rows, err := conn.Query(query, id)
	if err != nil {
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
func PerformTransaction(from_user_id int, user_to string, filename string) (int, error) {
	var id int
	to_user_id, err := GetUserID(user_to)
	if err != nil {
		return id, err
	}
	if AreMutualFriends(from_user_id, to_user_id) {
		query := `INSERT INTO transfer (from_user, to_user, filename) VALUES ($1,$2,$3) RETURNING id`
		if err := conn.QueryRow(query, from_user_id, to_user_id, filename).Scan(&id); err != nil {
			return id, err
		}
		return id, nil
	}
	return id, custom.ALREADYFRIENDS
}

/*
Retrieves the key from the database that is used to encrypt the file
*/
func RetrievePublicKey(id int) ([]byte, error) {
	var key []byte
	query := `SELECT key FROM publickey WHERE id = $1`
	if err := conn.QueryRow(query, id).Scan(&key); err != nil {
		return nil, err
	}

	return key, nil
}

/*
Inserts the public key into the userkey table
*/
func InsertPublicKey(id int, key []byte) error {
	query := `INSERT INTO publickey (id, key) VALUES ($1, $2)`
	if _, err := conn.Exec(query, id, key); err != nil {
		return err
	}

	return nil
}

/*
Inserts the encrypted nonce and key into rsa table
*/
func InsertRSA(nonce []byte, key []byte, id int) error {
	query := `INSERT INTO rsa (id, nonce, key) VALUES ($1, $2, $3)`
	if _, err := conn.Exec(query, id, nonce, key); err != nil {
		return err
	}

	return nil
}

/*
Returns the RSA encrypted nonce and key
*/
func RetrieveRSA(user int, file string) ([]byte, []byte, error) {
	var nonce, key []byte
	query := `
	SELECT nonce, key 
	FROM rsa 
	INNER JOIN transfer ON transfer.id=rsa.id

	WHERE to_user=$1 AND filename=$2`
	if err := conn.QueryRow(query, user, file).Scan(&nonce, &key); err != nil {
		return nil, nil, err
	}

	return nonce, key, nil
}
