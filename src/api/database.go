package database

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jackc/pgx"
)

type Sql struct {
	Host     string
	Port     uint16
	Database string
	User     string
	Password string
}

var conn *pgx.Conn
var connMutex sync.Mutex

// Create target connection for the database
func GetConn() (*pgx.Conn, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	if conn != nil {
		return conn, nil
	}
	home, _ := os.UserHomeDir()
	directory := filepath.Join(home, "Rapid/.sql.json")

	var sql Sql
	if err := parseSqlFile(&sql, directory); err != nil {
		fmt.Println(err)
	}
	connConfig := pgx.ConnConfig{
		Host:     sql.Host,
		Port:     sql.Port,
		Database: sql.Database,
		User:     sql.User,
		Password: sql.Password,
	}

	if newConn, err := pgx.Connect(connConfig); err != nil {
		return nil, fmt.Errorf("Failed to connect: %v", err)
	}

	conn = newConn
	return conn, nil
}

func parseSqlFile(sql *Sql, path string) error {
	if data, err := os.ReadFile(path); err != nil {
		return err
	}
	if err = json.Unmarshal(data, &sql); err != nil {
		return err
	}
	return nil
}

// Inits all of the tables for the database
func InitializeDatabase() error {
	var content embed.FS
	path, _ := content.ReadFile("database.sql")
	if conn, err := GetConn(); err != nil {
		return fmt.Errorf("Error connecting to the database: %v", err)
	}

	// Execute the SQL file
	if _, err = conn.Exec(string(path)); err != nil {
		return fmt.Errorf("Error executing SQL file: %v", err)
	}
	return nil
}
