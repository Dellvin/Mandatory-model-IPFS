package storage

import (
	"fmt"

	"github.com/jackc/pgx"
)

type Accum struct {
	ID           string
	WitnessLevel string
	WitnessDep   string
}

func CreateTableAccum(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "accum"(
ID TEXT PRIMARY KEY ,
witness_level TEXT,
witness_dep TEXT)`).Scan()
}

func GetWitness(conn *pgx.ConnPool, id string) (Accum, error) {
	var accum Accum
	err := conn.QueryRow("SELECT id, witness_level, witness_dep FROM accum WHERE id = $1;", id).Scan(&accum.ID, &accum.WitnessLevel, &accum.WitnessDep)

	if err == pgx.ErrNoRows {
		return Accum{}, err
	} else if err != nil {
		return Accum{}, fmt.Errorf("failed to Scan: %w", err)
	}

	return accum, nil
}

func SetAccum(conn *pgx.ConnPool, accum Accum) error {
	err := conn.QueryRow("INSERT INTO accum (id, witness_level, witness_dep) VALUES ($1, $2, $3)", accum.ID, accum.WitnessLevel, accum.WitnessDep).
		Scan(&accum.ID, &accum.WitnessLevel, &accum.WitnessDep)

	if err == pgx.ErrNoRows {
		return err
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}
