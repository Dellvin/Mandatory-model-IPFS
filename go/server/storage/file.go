package storage

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx"
)

type File struct {
	ID      int64
	Name    string
	IpfsKey string
	UserID  int
}

func CreateTableFile(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "file"(
id SERIAL PRIMARY KEY ,
name TEXT NOT NULL,
ipfs_key TEXT NOT NULL UNIQUE,
user_id int                               
)`).Scan()
}

func AddFile(conn *pgx.ConnPool, file File) error {
	err := conn.QueryRow("INSERT INTO file (name, ipfs_key, user_id) VALUES ($1, $2, $3, $4)", file.Name, file.IpfsKey, file.UserID).
		Scan(&file.ID, &file.Name, &file.IpfsKey, &file.UserID)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}

func GetFile(conn *pgx.ConnPool, userID int) (File, error) {
	var file File
	err := conn.QueryRow("SELECT * from file WHERE userID = $1", userID).Scan(&file.ID, &file.Name, &file.IpfsKey, &file.UserID)
	if err == pgx.ErrNoRows {
		return File{}, err
	} else if err != nil {
		return File{}, fmt.Errorf("failed to Scan: %w", err)
	}

	return file, nil
}

func GetAccessedFiles(conn *pgx.ConnPool, user User) ([]File, error) {
	var files []File
	rows, err := conn.Query(`SELECT f.id, f.name, f.ipfs_key, f.user_id from "file" as f
INNER JOIN "user" as u on f.user_id = u.id
WHERE u.level<=$1 AND u.department=$2`, user.Level, user.Department)
	if err != nil {
		return nil, fmt.Errorf("failed to Query: %w", err)
	}

	for rows.Next() {
		var file File
		if err = rows.Scan(&file.ID, &file.Name, &file.IpfsKey, &file.UserID); err != nil {
			return nil, fmt.Errorf("failed to Scan: %w", err)
		}

		files = append(files, file)
	}

	return files, nil
}
