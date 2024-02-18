package storage

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx"
)

type Witness struct {
	ID           string
	WitnessLevel string
	WitnessDep   string
}

func CreateTableWitness(conn *pgx.ConnPool) error {
	return conn.QueryRow(`
CREATE TABLE IF NOT EXISTS "witness"(
ID TEXT PRIMARY KEY ,
witness_level TEXT,
witness_dep TEXT)`).Scan()
}

func GetWitness(conn *pgx.ConnPool, id string) (Witness, error) {
	var witness Witness
	err := conn.QueryRow("SELECT id, witness_level, witness_dep FROM witness WHERE id = $1;", id).Scan(&witness.ID, &witness.WitnessLevel, &witness.WitnessDep)

	if err == pgx.ErrNoRows {
		return Witness{}, err
	} else if err != nil {
		return Witness{}, fmt.Errorf("failed to Scan: %w", err)
	}

	return witness, nil
}

func SetWitness(conn *pgx.ConnPool, witness Witness) error {
	err := conn.QueryRow("INSERT INTO witness (id, witness_level, witness_dep) VALUES ($1, $2, $3)", witness.ID, witness.WitnessLevel, witness.WitnessDep).
		Scan(&witness.ID, &witness.WitnessLevel, &witness.WitnessDep)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to Scan: %w", err)
	}

	return nil
}
