package storage

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
)

type AbeAuth struct {
	ID        string
	LevelAuth string
	DepAuth   string
}

func CreateTableAbe(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "auth"(
id TEXT,
level TEXT,
dep TEXT)`).Scan()

}

func GetAbe(conn *pgx.ConnPool, id string) (AbeAuth, error) {
	var abe AbeAuth
	err := conn.QueryRow("SELECT id, level, dep FROM auth WHERE id = $1;", id).Scan(&abe.ID, &abe.LevelAuth, &abe.DepAuth)

	if err == pgx.ErrNoRows {
		return AbeAuth{}, err
	} else if err != nil {
		return AbeAuth{}, fmt.Errorf("failed to Scan: %w", err)
	}

	return abe, nil
}

func SetAbe(conn *pgx.ConnPool, abe AbeAuth) error {
	err := conn.QueryRow("INSERT INTO auth (id, level, dep) VALUES ($1, $2, $3)", abe.ID, abe.LevelAuth, abe.DepAuth).
		Scan(&abe.ID, &abe.LevelAuth, &abe.DepAuth)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}
