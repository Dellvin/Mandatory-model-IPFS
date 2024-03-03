package storage

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"math"
	"math/big"
)

type User struct {
	ID         int
	PK         string
	Department int
	Level      int
}

func CreateTableUser(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "user"(
id SERIAL PRIMARY KEY ,
pk TEXT NOT NULL UNIQUE,
department int,
level int                                 
)`).Scan()
}

func CheckUser(conn *pgx.ConnPool, id int, pk string) error {
	var user User
	err := conn.QueryRow(`SELECT id FROM "user" WHERE id = $1 AND pk = $2;`, id, pk).Scan(&user.ID)

	if err == pgx.ErrNoRows {
		return err
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}

func GetUser(conn *pgx.ConnPool, id int, pk string) (User, error) {
	var user User
	err := conn.QueryRow(`SELECT * FROM "user" WHERE id = $1 AND pk = $2;`, id, pk).Scan(&user.ID, &user.PK, &user.Department, &user.Level)

	if err == pgx.ErrNoRows {
		return User{}, err
	} else if err != nil {
		return User{}, fmt.Errorf("failed to Scan: %w", err)
	}

	return user, nil
}

func GetAll(conn *pgx.ConnPool) ([]User, error) {
	var users []User
	rows, err := conn.Query(`SELECT * FROM "user";`)
	if err != nil {
		return nil, fmt.Errorf("failed to Query: %w", err)
	}

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.PK, &user.Department, &user.Level); err != nil {
			return nil, fmt.Errorf("failed to Scan: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func AddUser(conn *pgx.ConnPool, dep, level int) (User, error) {
	var user User
	for i := 0; i < 1000; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(int64(math.MaxInt64)))
		if err != nil {
			return User{}, fmt.Errorf("failed to rand.Int: %w", err)
		}
		pk := base64.StdEncoding.EncodeToString(nBig.Bytes())
		err = conn.QueryRow(`INSERT INTO "user" (pk, department, level) VALUES ($1, $2, $3)`, pk, dep, level).
			Scan(&user.ID, &user.PK, &user.Department, &user.Level)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			continue
		}

		err = conn.QueryRow(`SELECT * FROM "user" WHERE pk = $1;`, pk).Scan(&user.ID, &user.PK, &user.Department, &user.Level)
		if err == pgx.ErrNoRows {
			return User{}, err
		} else if err != nil {
			return User{}, fmt.Errorf("failed to SELECT: %w", err)
		}

		user.PK = pk

		return user, nil
	}

	return User{}, nil
}

func DeleteUser(conn *pgx.ConnPool, id int, pk string) error {
	var user User
	err := conn.QueryRow(`DELETE FROM "user" WHERE id = $1 OR pk = $2`, id, pk).Scan(&user.ID, &user.PK)

	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}
