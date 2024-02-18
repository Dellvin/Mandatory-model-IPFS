package storage

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
)

type Accumulator struct {
	Type  int
	Value int
	Accum string
}

func CreateTableAccumulator(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "accumulator"(
acc_type TEXT,
acc_value TEXT,
acc TEXT
)`).Scan()
}

func GetAccum(conn *pgx.ConnPool, accType, accValue string) (string, error) {
	var acc string
	err := conn.QueryRow("SELECT acc FROM accumulator WHERE acc_type = $1 AND acc_value=$2;", accType, accValue).Scan(&acc)

	if err == pgx.ErrNoRows {
		return "", err
	} else if err != nil {
		return "", fmt.Errorf("failed to Scan: %w", err)
	}

	return acc, nil
}

func SetAccum(conn *pgx.ConnPool, acc Accumulator) error {
	err := conn.QueryRow("INSERT INTO accumulator (acc_type, acc_value, acc) VALUES ($1, $2, $3)", acc.Type, acc.Value, acc.Accum).
		Scan(&acc.Type, &acc.Value, &acc.Accum)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}
