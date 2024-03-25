package database

import custom "github.com/Zaikoa/rapid/src/handling"

type Transaction struct {
	From_user string
	File_name string
}

/*
retrieves pending file transfer requests for a certain user
*/
func GetPendingTransfers(id int) ([]Transaction, error) {
	if id == 0 {
		return nil, custom.NewError("User must be logged in to use this method")
	}

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

	// dummy value that will store the id and is not checked (just used to check if no rows)
	var temp int // Temp cannot be null since sql does not support null primary keys
	err := row.Scan(&temp)
	if err != nil || temp == 0 {
		return false
	}

	return true
}

/*
Deletes a transaction record based on the given id
*/
func DeleteTransaction(id int, filename string) error {
	query := `DELETE FROM transfer WHERE to_user=$1 AND filename=$2`
	_, err := conn.Exec(query, id, filename)
	if err != nil {
		return err
	}
	return nil
}

/*
Enters the information of a transaction to the database for later referal
*/
func PerformTransaction(from_user_id int, user_to string, filename string) (bool, error) {
	to_user_id, _ := GetUserID(user_to)

	if AreMutualFriends(from_user_id, to_user_id) {
		query := `INSERT INTO transfer (from_user, to_user, filename) VALUES ($1,$2,$3)`
		_, err := conn.Exec(query, from_user_id, to_user_id, filename)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

/*
Retrieves the key from the database that is used to encrypt the file
*/
func RetrievePublicKey(id int) (string, error) {
	var key string
	query := `SELECT key FROM userkey WHERE users_id = $1`
	err := conn.QueryRow(query, id).Scan(&key)

	if err != nil {
		return "", err
	}

	return key, nil
}

/*
Inserts the public key into the userkey table
*/
func InsertPublicKey(id int, key string) error {
	query := `INSERT INTO userkey (users_id, key) VALUES ($1, $2)`
	_, err := conn.Exec(query, id, key)
	if err != nil {
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
	err := conn.QueryRow(query, string(nounce), string(key)).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

/*
Returns the RSA encrypted nounce and key
*/
func RetrieveRSA(id int) (string, string, error) {
	var nounce string
	var key string
	query := `SELECT nounce, key FROM rsa WHERE id=$1`
	err := conn.QueryRow(query, id).Scan(&nounce, &key)
	if err != nil {
		return "", "", err
	}

	return nounce, key, nil
}
